package generics

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gomods/athens/pkg/repo"
)

const (
	tmpRepoDir = "%s-%s" // owner-repo-ref
)

type genericFetcher struct {
	repoURI string
	version string
	dirName string
}

// NewGenericFetcher creates fetcher which uses go get tool to fetch sources
// returns path of directory containing vx.y.z.(zip|info|mod)
func NewGenericFetcher(repoURI, version string) (repo.Fetcher, error) {
	if !isVgoInstalled() {
		return nil, errors.New("vgo not installed")
	}

	if repoURI == "" {
		return nil, errors.New("invalid repository identifier")
	}

	return &genericFetcher{
		repoURI: repoURI,
		version: version,
	}, nil
}

// Fetch downloads the sources and returns path where it can be found
func (g *genericFetcher) Fetch() (string, error) {
	escapedURI := strings.Replace(g.repoURI, "/", "-", -1)
	repoDirName := fmt.Sprintf(tmpRepoDir, escapedURI, g.version)

	repoRoot, err := setupTmp(repoDirName)
	if err != nil {
		return "", err
	}
	g.dirName = repoRoot

	prepareStructure(repoRoot)

	dirName, err := getSources(repoRoot, g.repoURI, g.version)

	return dirName, err
}

// Clear removes all downloaded data
func (g *genericFetcher) Clear() error {
	if g.dirName == "" {
		return nil
	}

	return os.RemoveAll(g.dirName)
}

func isVgoInstalled() bool {
	cmd := exec.Command("vgo")
	err := cmd.Run() // this will fail even if vgo is installed

	if err != nil {
		_, ok := err.(*exec.ExitError) // if vgo is installed it will return ExitError, otherwise normal error
		return ok == true
	}

	return false
}

func setupTmp(repoDirName string) (string, error) {
	tmpDir := os.TempDir()
	path := filepath.Join(tmpDir, repoDirName)

	return path, os.MkdirAll(path, os.ModeDir|os.ModePerm)
}

// Hacky thing makes vgo not to complain
func prepareStructure(repoRoot string) error {
	// vgo expects go.mod file present with module statement or .go file with import comment
	gomodPath := filepath.Join(repoRoot, "go.mod")
	gomodContent := []byte("module \"mod\"")
	if err := ioutil.WriteFile(gomodPath, gomodContent, 0666); err != nil {
		return err
	}

	sourcePath := filepath.Join(repoRoot, "mod.go")
	sourceContent := []byte(`package mod // import "mod"`)
	return ioutil.WriteFile(sourcePath, sourceContent, 0666)
}

func getSources(repoRoot, repoURI, version string) (string, error) {
	version = strings.TrimPrefix(version, "@")
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	uri := strings.TrimSuffix(repoURI, "/")

	fullURI := fmt.Sprintf("%s@%s", uri, version)

	gopathEnv := fmt.Sprintf("GOPATH=%s", repoRoot)

	cmd := exec.Command("vgo", "get", fullURI)
	cmd.Env = os.Environ()
	cmd.Env = append([]string{gopathEnv}, cmd.Env...)
	cmd.Dir = repoRoot

	packagePath := filepath.Join(repoRoot, "src", "v", "cache", repoURI, "@v")

	err := cmd.Run()
	if err != nil && checkFiles(packagePath, version) == nil {
		// is some compilation error
		return packagePath, nil
	}

	return packagePath, err
}

func checkFiles(path, version string) error {
	if _, err := os.Stat(filepath.Join(path, version+".mod")); err != nil {
		return errors.New("go.mod not found")
	}

	if _, err := os.Stat(filepath.Join(path, version+".zip")); err != nil {
		return errors.New("zip package not found")
	}

	if _, err := os.Stat(filepath.Join(path, version+".info")); err != nil {
		return errors.New("info file not found")
	}

	return nil
}
