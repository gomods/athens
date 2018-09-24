package module

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
)

type goGetFetcher struct {
	fs           afero.Fs
	goBinaryName string
}

// NewGoGetFetcher creates fetcher which uses go get tool to fetch modules
func NewGoGetFetcher(goBinaryName string, fs afero.Fs) (Fetcher, error) {
	const op errors.Op = "module.NewGoGetFetcher"
	if err := validGoBinary(goBinaryName); err != nil {
		return nil, errors.E(op, err)
	}
	return &goGetFetcher{
		fs:           fs,
		goBinaryName: goBinaryName,
	}, nil
}

// Fetch downloads the sources from the go binary and returns the corresponding
// .info, .mod, and .zip files.
func (g *goGetFetcher) Fetch(ctx context.Context, mod, ver string) (*storage.Version, error) {
	const op errors.Op = "goGetFetcher.Fetch"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	// setup the GOPATH
	goPathRoot, err := afero.TempDir(g.fs, "", "athens")
	if err != nil {
		return nil, errors.E(op, err)
	}
	sourcePath := filepath.Join(goPathRoot, "src")
	modPath := filepath.Join(sourcePath, getRepoDirName(mod, ver))
	if err := g.fs.MkdirAll(modPath, os.ModeDir|os.ModePerm); err != nil {
		ClearFiles(g.fs, goPathRoot)
		return nil, errors.E(op, err)
	}

	// setup the module with barebones stuff
	if err := Dummy(g.fs, modPath); err != nil {
		ClearFiles(g.fs, goPathRoot)
		return nil, errors.E(op, err)
	}

	err = getSources(g.goBinaryName, g.fs, goPathRoot, modPath, mod, ver)
	if err != nil {
		ClearFiles(g.fs, goPathRoot)
		return nil, errors.E(op, err)
	}

	dr := newDiskRef(g.fs, goPathRoot, mod, ver)
	return dr.Read()
}

// Dummy Hacky thing makes vgo not to complain
func Dummy(fs afero.Fs, repoRoot string) error {
	const op errors.Op = "module.Dummy"
	// vgo expects go.mod file present with module statement or .go file with import comment
	gomodPath := filepath.Join(repoRoot, "go.mod")
	gomodContent := []byte("module mod")
	if err := afero.WriteFile(fs, gomodPath, gomodContent, 0666); err != nil {
		return errors.E(op, err)
	}
	sourcePath := filepath.Join(repoRoot, "mod.go")
	sourceContent := []byte("package mod")
	if err := afero.WriteFile(fs, sourcePath, sourceContent, 0666); err != nil {
		return errors.E(op, err)
	}
	return nil
}

// given a filesystem, gopath, repository root, module and version, runs 'vgo get'
// on module@version from the repoRoot with GOPATH=gopath, and returns a non-nil error if anything went wrong.
func getSources(goBinaryName string, fs afero.Fs, gopath, repoRoot, module, version string) error {
	const op errors.Op = "module.getSources"
	uri := strings.TrimSuffix(module, "/")

	// Check the output of fullURI
	fullURI := fmt.Sprintf("%s@%s", uri, version)
	cmd := downloadCmd(goBinaryName, fullURI, "mod", "download")
	cmd.Env = PrepareEnv(gopath)
	cmd.Dir = repoRoot
	o, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := fmt.Sprintf("%v : %s", err, o)
		// github quota exceeded
		if isLimitHit(o) {
			return errors.E(op, errMsg, errors.KindRateLimit)
		}
		// another error in the output
		return errors.E(op, errMsg)
	}
	// make sure the expected files exist
	encmod, err := paths.EncodePath(module)
	if err != nil {
		return errors.E(op, err)
	}
	packagePath := getPackagePath(gopath, encmod)
	err = checkFiles(fs, packagePath, version)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func downloadCmd(goBinaryName string, fullURI string, args ...string) *exec.Cmd {
	insecureSources := strings.Split(os.Getenv("PROXY_INSECURE_SOURCES"), ",")
	modSource := strings.Split(fullURI, "/")[0]
	insecureFlag := false

	for _, source := range insecureSources {
		if modSource == source {
			insecureFlag = true
			break
		}
	}

	if insecureFlag {
		args = append(args, "-insencure")
	}

	args = append(args, fullURI)
	return exec.Command(goBinaryName, args...)
}

// PrepareEnv will return all the appropriate
// environment variables for a Go Command to run
// successfully (such as GOPATH, GOCACHE, PATH etc)
func PrepareEnv(gopath string) []string {
	pathEnv := fmt.Sprintf("PATH=%s", os.Getenv("PATH"))
	httpProxy := fmt.Sprintf("HTTP_PROXY=%s", os.Getenv("HTTP_PROXY"))
	httpsProxy := fmt.Sprintf("HTTPS_PROXY=%s", os.Getenv("HTTPS_PROXY"))
	noProxy := fmt.Sprintf("NO_PROXY=%s", os.Getenv("NO_PROXY"))
	gopathEnv := fmt.Sprintf("GOPATH=%s", gopath)
	cacheEnv := fmt.Sprintf("GOCACHE=%s", filepath.Join(gopath, "cache"))
	disableCgo := "CGO_ENABLED=0"
	enableGoModules := "GO111MODULE=on"
	cmdEnv := []string{pathEnv, gopathEnv, cacheEnv, disableCgo, enableGoModules, httpProxy, httpsProxy, noProxy}

	// add Windows specific ENV VARS
	if runtime.GOOS == "windows" {
		cmdEnv = append(cmdEnv, fmt.Sprintf("USERPROFILE=%s", os.Getenv("USERPROFILE")))
		cmdEnv = append(cmdEnv, fmt.Sprintf("SystemRoot=%s", os.Getenv("SystemRoot")))
		cmdEnv = append(cmdEnv, fmt.Sprintf("ALLUSERSPROFILE=%s", os.Getenv("ALLUSERSPROFILE")))
		cmdEnv = append(cmdEnv, fmt.Sprintf("HOMEDRIVE=%s", os.Getenv("HOMEDRIVE")))
		cmdEnv = append(cmdEnv, fmt.Sprintf("HOMEPATH=%s", os.Getenv("HOMEPATH")))
	}

	return cmdEnv
}

func checkFiles(fs afero.Fs, path, version string) error {
	const op errors.Op = "module.checkFiles"
	if _, err := fs.Stat(filepath.Join(path, version+".mod")); err != nil {
		return errors.E(op, fmt.Sprintf("%s.mod not found in %s", version, path))
	}

	if _, err := fs.Stat(filepath.Join(path, version+".zip")); err != nil {
		return errors.E(op, fmt.Sprintf("%s.mod not found in %s", version, path))
	}

	if _, err := fs.Stat(filepath.Join(path, version+".info")); err != nil {
		return errors.E(op, fmt.Sprintf("%s.mod not found in %s", version, path))
	}

	return nil
}

func isLimitHit(o []byte) bool {
	return bytes.Contains(o, []byte("403 response from api.github.com"))
}

// getRepoDirName takes a raw repository URI and a version and creates a directory name that the
// repository contents can be put into
func getRepoDirName(repoURI, version string) string {
	escapedURI := strings.Replace(repoURI, "/", "-", -1)
	return fmt.Sprintf("%s-%s", escapedURI, version)
}

// getPackagePath returns the path to the module cache given the gopath and module name
func getPackagePath(gopath, module string) string {
	return filepath.Join(gopath, "pkg", "mod", "cache", "download", module, "@v")
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
