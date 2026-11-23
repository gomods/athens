package module

import (
	"bytes"
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
)

type goGetFetcher struct {
	fs           afero.Fs
	goBinaryName string
	envVars      []string
	gogetDir     string
}

type goModule struct {
	Path     string `json:"path"`     // module path
	Version  string `json:"version"`  // module version
	Error    string `json:"error"`    // error loading module
	Info     string `json:"info"`     // absolute path to cached .info file
	GoMod    string `json:"goMod"`    // absolute path to cached .mod file
	Zip      string `json:"zip"`      // absolute path to cached .zip file
	Dir      string `json:"dir"`      // absolute path to cached source root directory
	Sum      string `json:"sum"`      // checksum for path, version (as in go.sum)
	GoModSum string `json:"goModSum"` // checksum for go.mod (as in go.sum)
}

// NewGoGetFetcher creates fetcher which uses go get tool to fetch modules.
func NewGoGetFetcher(goBinaryName, gogetDir string, envVars []string, fs afero.Fs) (Fetcher, error) {
	const op errors.Op = "module.NewGoGetFetcher"
	if err := validGoBinary(goBinaryName); err != nil {
		return nil, errors.E(op, err)
	}
	return &goGetFetcher{
		fs:           fs,
		goBinaryName: goBinaryName,
		envVars:      envVars,
		gogetDir:     gogetDir,
	}, nil
}

// Fetch downloads the sources from the go binary and returns the corresponding
// .info, .mod, and .zip files.
func (g *goGetFetcher) Fetch(ctx context.Context, mod, ver string) (_ *storage.Version, err error) {
	const op errors.Op = "goGetFetcher.Fetch"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	// setup the GOPATH
	goPathRoot, err := afero.TempDir(g.fs, g.gogetDir, "athens")
	if err != nil {
		return nil, errors.E(op, err)
	}
	// On success, ownership of goPathRoot transfers to the returned zipReadCloser,
	// which clears it on Close. On every error return, clean it up here.
	defer func() {
		if err != nil {
			_ = clearFiles(g.fs, goPathRoot)
		}
	}()

	sourcePath := filepath.Join(goPathRoot, "src")
	modPath := filepath.Join(sourcePath, getRepoDirName(mod, ver))
	if err := g.fs.MkdirAll(modPath, os.ModeDir|os.ModePerm); err != nil {
		return nil, errors.E(op, err)
	}

	m, err := downloadModule(
		ctx,
		g.goBinaryName,
		g.envVars,
		goPathRoot,
		modPath,
		mod,
		ver,
	)
	if err != nil {
		return nil, errors.E(op, err)
	}

	var storageVer storage.Version
	storageVer.Semver = m.Version
	info, err := afero.ReadFile(g.fs, m.Info)
	if err != nil {
		return nil, errors.E(op, err)
	}
	storageVer.Info = info

	gomod, err := afero.ReadFile(g.fs, m.GoMod)
	if err != nil {
		return nil, errors.E(op, err)
	}
	storageVer.Mod = gomod

	zipMD5, err := func() ([]byte, error) {
		// Perform in a separate function to ensure file is closed
		zipForChecksum, err := g.fs.Open(m.Zip)
		if err != nil {
			return nil, errors.E(op, err)
		}
		defer zipForChecksum.Close()

		//nolint:gosec
		hash := md5.New()
		if _, err := io.Copy(hash, zipForChecksum); err != nil {
			return nil, errors.E(op, err)
		}

		return hash.Sum(nil), nil
	}()
	if err != nil {
		return nil, err
	}

	zip, err := g.fs.Open(m.Zip)
	if err != nil {
		return nil, errors.E(op, err)
	}
	// note: don't close zip here so that the caller can read directly from disk.
	//
	// if we close, then the caller will panic, and the alternative to make this work is
	// that we read into memory and return an io.ReadCloser that reads out of memory
	storageVer.Zip = &zipReadCloser{zip, g.fs, goPathRoot}
	storageVer.ZipMD5 = zipMD5

	return &storageVer, nil
}

// given a filesystem, gopath, repository root, module and version, runs 'go mod download -json'
// on module@version from the repoRoot with GOPATH=gopath, and returns a non-nil error if anything went wrong.
func downloadModule(
	ctx context.Context,
	goBinaryName string,
	envVars []string,
	gopath,
	repoRoot,
	module,
	version string,
) (goModule, error) {
	const op errors.Op = "module.downloadModule"

	uri := strings.TrimSuffix(module, "/")
	fullURI := fmt.Sprintf("%s@%s", uri, version)

	cmd := exec.CommandContext(ctx, goBinaryName, "mod", "download", "-json", fullURI)
	cmd.Env = prepareEnv(gopath, envVars)
	cmd.Dir = repoRoot
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	runErr := cmd.Run()

	// `go mod download -json` reports the outcome of the module as a JSON object
	// on stdout whose Error field is set iff the download failed; the process
	// exit code merely mirrors that field. We therefore classify from the JSON
	// output, treating the exit code as advisory only.
	var m goModule
	jsonErr := json.NewDecoder(stdout).Decode(&m)

	switch {
	case jsonErr == nil && m.Error == "":
		// go reported a successful download.
		return m, nil
	case jsonErr == nil:
		// go reported a per-module failure (unknown revision, no go-import meta
		// tags, invalid version, ...). Never a 5xx — see goGetErrKind.
		return goModule{}, errors.E(op, m.Error, goGetErrKind(m.Error))
	case runErr != nil:
		// The process failed and left no parseable JSON on stdout (the
		// diagnostic went only to stderr). Classify from stderr.
		err := fmt.Errorf("%w: %s", runErr, stderr)
		return goModule{}, errors.E(op, err, goGetErrKind(stderr.String()))
	default:
		// The process exited zero yet produced no parseable module metadata —
		// nothing trustworthy to classify from. Surface as not-found so the go
		// client falls back to ,direct instead of aborting on a fatal 500.
		err := fmt.Errorf("go mod download produced no module metadata: %w: %s", jsonErr, stderr)
		return goModule{}, errors.E(op, err, errors.KindNotFound)
	}
}

func isLimitHit(o string) bool {
	return strings.Contains(o, "403 response from api.github.com")
}

// goGetErrKind classifies a `go mod download` / `go list` failure message into
// an errors.Kind. Any failure to fetch a specific module@version is reported as
// KindNotFound (404), never KindUnexpected (500): when resolving `pkg@version`
// the go command probes candidate module paths longest-to-shortest and treats
// *any* non-404/410 status (including 500) as fatal, while the GOPROXY
// `,direct` fallback only triggers on 404/410. Rate-limit responses keep their
// dedicated kind so callers can back off.
// See https://github.com/golang/go/issues/30134.
func goGetErrKind(msg string) int {
	if isLimitHit(msg) {
		return errors.KindRateLimit
	}

	return errors.KindNotFound
}

// getRepoDirName takes a raw repository URI and a version and creates a directory name that the
// repository contents can be put into.
func getRepoDirName(repoURI, version string) string {
	escapedURI := strings.ReplaceAll(repoURI, "/", "-")
	return fmt.Sprintf("%s-%s", escapedURI, version)
}

func validGoBinary(name string) error {
	const op errors.Op = "module.validGoBinary"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := exec.CommandContext(ctx, name).Run()
	eErr := &exec.ExitError{}
	if err != nil && !errors.AsErr(err, &eErr) {
		return errors.E(op, err)
	}
	return nil
}
