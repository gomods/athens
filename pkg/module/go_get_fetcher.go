package module

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	pkgerrors "github.com/pkg/errors"
	"github.com/spf13/afero"
)

var (
	// ErrLimitExceeded signals that github.com refused to serve the request due to exceeded quota
	ErrLimitExceeded = errors.New("github limit exceeded")
)

type goGetFetcher struct {
	fs           afero.Fs
	repoURI      string
	version      string
	goBinaryName string
}

// NewGoGetFetcher creates fetcher which uses go get tool to fetch modules
func NewGoGetFetcher(goBinaryName string, fs afero.Fs, repoURI, version string) (Fetcher, error) {
	if repoURI == "" {
		return nil, errors.New("invalid repository identifier")
	}

	return &goGetFetcher{
		fs:           fs,
		repoURI:      repoURI,
		version:      version,
		goBinaryName: goBinaryName,
	}, nil
}

// Fetch downloads the sources and returns path where it can be found. This function will
// always return a ref, even if it returns a non-nil error.
//
// If an error was returned, the returned ref's Read method will return an error, but you
// should always call ref.Clear() on the returned ref
func (g *goGetFetcher) Fetch(mod, ver string) (Ref, error) {
	var ref Ref
	ref = noopRef{}

	// setup the GOPATH
	goPathRoot, err := afero.TempDir(g.fs, "", "athens")
	if err != nil {
		// TODO: return a ref for cleaning up the goPathRoot
		// https://github.com/gomods/athens/issues/329
		return ref, err
	}
	sourcePath := filepath.Join(goPathRoot, "src")
	modPath := filepath.Join(sourcePath, getRepoDirName(g.repoURI, g.version))
	if err := g.fs.MkdirAll(modPath, os.ModeDir|os.ModePerm); err != nil {
		// TODO: return a ref for cleaning up the goPathRoot
		// https://github.com/gomods/athens/issues/329
		return ref, err
	}

	// setup the module with barebones stuff
	if err := prepareStructure(g.fs, modPath); err != nil {
		// TODO: return a ref for cleaning up the goPathRoot
		// https://github.com/gomods/athens/issues/329
		return ref, err
	}

	cachePath, err := getSources(g.goBinaryName, g.fs, goPathRoot, modPath, mod, ver)
	if err != nil {
		// TODO: return a ref that cleans up the goPathRoot
		// https://github.com/gomods/athens/issues/329
		return newDiskRef(g.fs, cachePath, ver), err
	}
	// TODO: make sure this ref also cleans up the goPathRoot
	// https://github.com/gomods/athens/issues/329
	ref = newDiskRef(g.fs, cachePath, ver)

	return ref, err
}

// Hacky thing makes vgo not to complain
func prepareStructure(fs afero.Fs, repoRoot string) error {
	// vgo expects go.mod file present with module statement or .go file with import comment
	gomodPath := filepath.Join(repoRoot, "go.mod")
	gomodContent := []byte("module mod")
	if err := afero.WriteFile(fs, gomodPath, gomodContent, 0666); err != nil {
		return err
	}

	sourcePath := filepath.Join(repoRoot, "mod.go")
	sourceContent := []byte(`package mod // import "mod"`)
	return afero.WriteFile(fs, sourcePath, sourceContent, 0666)
}

// given a filesystem, gopath, repository root, module and version, runs 'vgo get'
// on module@version from the repoRoot with GOPATH=gopath, and returns the location
// of the module cache. returns a non-nil error if anything went wrong. always returns
// the location of the module cache so you can delete it if necessary
func getSources(goBinaryName string, fs afero.Fs, gopath, repoRoot, module, version string) (string, error) {
	version = strings.TrimPrefix(version, "@")
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	uri := strings.TrimSuffix(module, "/")

	fullURI := fmt.Sprintf("%s@%s", uri, version)

	gopathEnv := fmt.Sprintf("GOPATH=%s", gopath)
	cacheEnv := fmt.Sprintf("GOCACHE=%s", filepath.Join(gopath, "cache"))
	disableCgo := "CGO_ENABLED=0"

	cmd := exec.Command(goBinaryName, "get", fullURI)
	// PATH is needed for vgo to recognize vcs binaries
	// this breaks windows.
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), gopathEnv, cacheEnv, disableCgo}
	cmd.Dir = repoRoot

	packagePath := filepath.Join(gopath, "src", "mod", "cache", "download", module, "@v")

	o, err := cmd.CombinedOutput()
	if err != nil {
		// github quota exceeded
		if isLimitHit(o) {
			return packagePath, pkgerrors.WithMessage(err, "github API limit hit")
		}
		// another error in the output
		return packagePath, err
	}
	// make sure the expected files exist
	if err := checkFiles(fs, packagePath, version); err != nil {
		return packagePath, err
	}

	return packagePath, nil
}

func checkFiles(fs afero.Fs, path, version string) error {
	if _, err := fs.Stat(filepath.Join(path, version+".mod")); err != nil {
		return pkgerrors.WithMessage(err, fmt.Sprintf("%s.mod not found in %s", version, path))
	}

	if _, err := fs.Stat(filepath.Join(path, version+".zip")); err != nil {
		return pkgerrors.WithMessage(err, fmt.Sprintf("%s.zip not found in %s", version, path))
	}

	if _, err := fs.Stat(filepath.Join(path, version+".info")); err != nil {
		return pkgerrors.WithMessage(err, fmt.Sprintf("%s.info not found in %s", version, path))
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
