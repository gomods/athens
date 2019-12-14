package module

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/gobuffalo/envy"
	"github.com/stretchr/testify/suite"
)

type ModuleSuite struct {
	suite.Suite
	goBinaryPath   string
	env            []string
	tmpDir         string
	goModCache     string
	sampleRepoPath string
}

type catalogRes struct {
	Modules []struct {
		Module  string `json:"module"`
		Version string `json:"version"`
	} `json:"modules"`
}

func (m *ModuleSuite) SetupSuite() {
	var err error
	m.tmpDir, err = mkTemp()
	if err != nil {
		m.Fail("Failed to make temp dir", err)
	}

	m.goModCache = path.Join(m.tmpDir, "pkg", "mod")
	m.sampleRepoPath = path.Join(m.tmpDir, "happy-path")
	m.goBinaryPath = envy.Get("GO_BINARY_PATH", "go")

	athensBin, err := buildAthens(m.goBinaryPath, m.tmpDir, m.env)
	if err != nil {
		m.Fail("Failed to build athens ", err)
	}
	stopAthens() // in case a dangling instance was around.
	// ignoring error as if no athens is running it fails.
	runAthensAndWait(athensBin, m.getEnv())
	setupTestRepo(m.sampleRepoPath)
}

func (m *ModuleSuite) TearDownSuite() {
	err := stopAthens()
	if err != nil {
		m.Fail("Failed to stop athens", err)
	}
	chmodR(m.tmpDir, 0777)
	os.RemoveAll(m.tmpDir)
}

func TestE2E(t *testing.T) {
	suite.Run(t, &ModuleSuite{})
}

func (m *ModuleSuite) SetupTest() {
	chmodR(m.goModCache, 0777)
	os.RemoveAll(m.goModCache)
}

func (m *ModuleSuite) TestNoGoProxy() {
	cmd := exec.Command("go", "run", ".")
	cmd.Env = m.env
	cmd.Dir = m.sampleRepoPath

	err := cmd.Run()
	if err != nil {
		m.Fail("go run failed on test repo", err)
	}
}

func (m *ModuleSuite) TestGoProxy() {
	cmd := exec.Command("go", "run", ".")
	cmd.Env = m.getEnvGoProxy(m.tmpDir)
	cmd.Dir = m.sampleRepoPath
	err := cmd.Run()
	if err != nil {
		m.Fail("go run failed on test repo", err)
	}
	resp, err := http.Get("http://localhost:3000/catalog")
	if err != nil {
		m.Fail("failed to read catalog", err)
	}

	var catalog catalogRes
	err = json.NewDecoder(resp.Body).Decode(&catalog)
	if err != nil {
		m.Fail("failed to decode catalog res", err)
	}
	m.Assert().Equal(len(catalog.Modules), 1)
	m.Assert().Equal(catalog.Modules[0].Module, "github.com/athens-artifacts/no-tags")
}

func (m *ModuleSuite) TestWrongGoProxy() {
	cmd := exec.Command("go", "run", ".")
	cmd.Env = m.getEnvWrongGoProxy(m.tmpDir)
	cmd.Dir = m.sampleRepoPath
	err := cmd.Run()
	m.Assert().NotNil(err, "Wrong proxy should fail")
}

func (m *ModuleSuite) getEnv() []string {
	res := []string{
		fmt.Sprintf("GOPATH=%s", m.tmpDir),
		"GO111MODULE=on",
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("GOCACHE=%s", filepath.Join(m.tmpDir, "cache")),
	}
	return res
}

func (m *ModuleSuite) getEnvGoProxy(gopath string) []string {
	res := m.getEnv()
	res = append(res, "GOPROXY=http://localhost:3000")
	return res
}

func (m *ModuleSuite) getEnvWrongGoProxy(gopath string) []string {
	res := m.getEnv()
	res = append(res, "GOPROXY=http://localhost:3001")
	return res
}

func setupTestRepo(repoPath string) {
	os.RemoveAll(repoPath)
	cmd := exec.Command("git",
		"clone",
		"https://github.com/athens-artifacts/happy-path.git",
		repoPath)

	cmd.Run()
}

func mkTemp() (string, error) {
	res, err := ioutil.TempDir("/tmp", "gopath")
	if err != nil {
		return "", fmt.Errorf("Failed to create tmpdir in /tmp %v", err)
	}
	return res, nil
}

func buildAthens(goBin string, destPath string, env []string) (string, error) {
	target := path.Join(destPath, "athens-proxy")
	binFolder, err := filepath.Abs("../cmd/proxy")
	if err != nil {
		return "", fmt.Errorf("Failed to get athens source path %v", err)
	}

	cmd := exec.Command(goBin, "build", "-o", target, binFolder)
	cmd.Env = env
	err = cmd.Run()
	return target, err
}

func stopAthens() error {
	cmd := exec.Command("pkill", "athens-proxy")
	err := cmd.Run()
	return err
}

func runAthensAndWait(athensBin string, env []string) error {
	cmd := exec.Command(athensBin)
	cmd.Env = env

	go func() {
		cmd.Run()
	}()

	ticker := time.NewTicker(time.Second)
	timeout := time.After(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			resp, _ := http.Get("http://localhost:3000/readyz")
			if resp.StatusCode == 200 {
				return nil
			}
		case <-timeout:
			return fmt.Errorf("Failed to run athens")
		}
	}
}

func chmodR(path string, mode os.FileMode) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			os.Chmod(name, mode)
		}
		return err
	})
}
