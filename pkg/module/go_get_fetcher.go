package module

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
)

type goGetFetcher struct {
	fs afero.Fs

	goTool
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

// NewGoGetFetcher creates fetcher which uses go get tool to fetch modules
func NewGoGetFetcher(goBinaryName string, goProxy string, fs afero.Fs) (Fetcher, error) {
	const op errors.Op = "module.NewGoGetFetcher"
	if err := validGoBinary(goBinaryName); err != nil {
		return nil, errors.E(op, err)
	}
	return &goGetFetcher{
		fs: fs,
		goTool: goTool{
			goBin:   goBinaryName,
			goProxy: goProxy,
		},
	}, nil
}

// Fetch downloads the sources from the go binary and returns the corresponding
// .info, .mod, and .zip files.
func (g *goGetFetcher) Fetch(ctx context.Context, mod, ver string) (*storage.Version, error) {
	const op errors.Op = "goGetFetcher.Fetch"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	m, cleanFunc, err := downloadModule(g.goTool, g.fs, mod, ver)
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

	zip, err := g.fs.Open(m.Zip)
	if err != nil {
		return nil, errors.E(op, err)
	}
	// note: don't close zip here so that the caller can read directly from disk.
	//
	// if we close, then the caller will panic, and the alternative to make this work is
	// that we read into memory and return an io.ReadCloser that reads out of memory
	storageVer.Zip = &zipReadCloser{zip, g.fs, cleanFunc}

	return &storageVer, nil
}

// given a filesystem, gopath, repository root, module and version, runs 'go mod download -json'
// on module@version from the repoRoot with GOPATH=gopath, and returns a non-nil error if anything went wrong.
func downloadModule(goTool goTool, fs afero.Fs, module, version string,
) (m *goModule, cleanFunc runtimeClean, err error) {
	const op errors.Op = "module.downloadModule"
	uri := strings.TrimSuffix(module, "/")
	fullURI := fmt.Sprintf("%s@%s", uri, version)
	repoRoot := filepath.Join("src", getRepoDirName(module, version))
	runtime, err := prepareRuntime(fs, goTool, repoRoot)
	if err != nil {
		return nil, nil, errors.E(op, err)
	}
	defer func() {
		if err != nil {
			runtime.clean()
		}
	}()

	m = &goModule{}
	err = runtime.run("mod", "download", "-json", fullURI)
	if err != nil {
		err = fmt.Errorf("%v: %s", err, runtime.stderr)

		if jsonErr := json.NewDecoder(runtime.stdout).Decode(m); jsonErr != nil {
			return nil, nil, errors.E(op, err)
		}
		// github quota exceeded
		if isLimitHit(m.Error) {
			return nil, nil, errors.E(op, m.Error, errors.KindRateLimit)
		}
		return nil, nil, errors.E(op, m.Error, errors.KindNotFound)
	}
	if err = json.NewDecoder(runtime.stdout).Decode(m); err != nil {
		return nil, nil, errors.E(op, err)
	}
	if m.Error != "" {
		return nil, nil, errors.E(op, m.Error)
	}
	return m, runtime.clean, nil
}

func isLimitHit(o string) bool {
	return strings.Contains(o, "403 response from api.github.com")
}

// getRepoDirName takes a raw repository URI and a version and creates a directory name that the
// repository contents can be put into
func getRepoDirName(repoURI, version string) string {
	escapedURI := strings.Replace(repoURI, "/", "-", -1)
	return fmt.Sprintf("%s-%s", escapedURI, version)
}

func validGoBinary(name string) error {
	const op errors.Op = "module.validGoBinary"
	err := exec.Command(name).Run()
	_, ok := err.(*exec.ExitError)
	if err != nil && !ok {
		return errors.E(op, err)
	}
	return nil
}
