package module

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

var (
	// ErrLimitExceeded signals that github.com refused to serve the request due to exceeded quota
	ErrLimitExceeded = errors.New("github limit exceeded")
)

type goGetFetcher struct {
	fs      afero.Fs
	repoURI string
	version string
}

// NewGoGetFetcher creates fetcher which uses go get tool to fetch modules
func NewGoGetFetcher(fs afero.Fs, repoURI, version string) (Fetcher, error) {
	if repoURI == "" {
		return nil, errors.New("invalid repository identifier")
	}

	return &goGetFetcher{
		fs:      fs,
		repoURI: repoURI,
		version: version,
	}, nil
}

// Fetch downloads the sources and returns path where it can be found
func (g *goGetFetcher) Fetch(mod, ver string) (Ref, error) {
	repoDirName := getRepoDirName(g.repoURI, g.version)

	gopath, repoRoot, err := setupTmp(g.fs, repoDirName)
	if err != nil {
		return nil, err
	}

	prepareStructure(g.fs, repoRoot)

	dirName, err := getSources(g.fs, gopath, repoRoot, g.repoURI, g.version)
	diskRef := newDiskRef(g.fs, dirName, ver)

	return diskRef, err
}

func setupTmp(fs afero.Fs, repoDirName string) (string, string, error) {
	gopathDir, err := afero.TempDir(fs, "", "")
	if err != nil {
		return "", "", err
	}

	path := filepath.Join(gopathDir, "src", repoDirName)

	return gopathDir, path, fs.MkdirAll(path, os.ModeDir|os.ModePerm)
}

// Hacky thing makes vgo not to complain
func prepareStructure(fs afero.Fs, repoRoot string) error {
	// vgo expects go.mod file present with module statement or .go file with import comment
	gomodPath := filepath.Join(repoRoot, "go.mod")
	gomodContent := []byte("module \"mod\"")
	if err := afero.WriteFile(fs, gomodPath, gomodContent, 0666); err != nil {
		return err
	}

	sourcePath := filepath.Join(repoRoot, "mod.go")
	sourceContent := []byte(`package mod // import "mod"`)
	return afero.WriteFile(fs, sourcePath, sourceContent, 0666)
}

func getSources(fs afero.Fs, gopath, repoRoot, repoURI, version string) (string, error) {
	version = strings.TrimPrefix(version, "@")
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	uri := strings.TrimSuffix(repoURI, "/")

	fullURI := fmt.Sprintf("%s@%s", uri, version)

	gopathEnv := fmt.Sprintf("GOPATH=%s", gopath)
	cacheEnv := fmt.Sprintf("GOCACHE=%s", filepath.Join(gopath, "cache"))
	disableCgo := "CGO_ENABLED=0"

	cmd := exec.Command("vgo", "get", fullURI)
	// PATH is needed for vgo to recognize vcs binaries
	// this breaks windows.
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), gopathEnv, cacheEnv, disableCgo}
	cmd.Dir = repoRoot

	packagePath := filepath.Join(gopath, "src", "mod", "cache", "download", repoURI, "@v")

	o, err := cmd.CombinedOutput()
	if err != nil {
		// github quota exceeded
		if isLimitHit(o) {
			return "", errors.New("github API limit hit")
		}
		// one or more of the expected files doesn't exist
		if err := checkFiles(fs, packagePath, version); err != nil {
			return "", err
		}
		// another error in the output
		return "", err
	}

	return packagePath, nil
}

func checkFiles(fs afero.Fs, path, version string) error {
	if _, err := fs.Stat(filepath.Join(path, version+".mod")); err != nil {
		return fmt.Errorf("%s.mod not found", version)
	}

	if _, err := fs.Stat(filepath.Join(path, version+".zip")); err != nil {
		return fmt.Errorf("%s.zip not found", version)
	}

	if _, err := fs.Stat(filepath.Join(path, version+".info")); err != nil {
		return fmt.Errorf("%s.info not found", version)
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
