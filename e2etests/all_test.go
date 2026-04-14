//go:build e2etests
// +build e2etests

package e2etests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gobuffalo/envy"
	"github.com/stretchr/testify/suite"
)

type E2eSuite struct {
	suite.Suite
	goBinaryPath   string
	env            []string
	goPath         string
	sampleRepoPath string
	stopAthens     context.CancelFunc
}

type catalogRes struct {
	Modules []struct {
		Module  string `json:"module"`
		Version string `json:"version"`
	} `json:"modules"`
}

func (m *E2eSuite) SetupSuite() {
	var err error
	m.goPath, err = os.MkdirTemp("/tmp", "gopath")
	if err != nil {
		m.Fail("Failed to make temp dir", err)
	}

	m.sampleRepoPath, err = os.MkdirTemp("/tmp", "repopath")
	if err != nil {
		m.Fail("Failed to make temp dir for sample repo", err)
	}

	m.goBinaryPath = envy.Get("GO_BINARY_PATH", "go")

	athensBin, err := buildAthens(m.goBinaryPath, m.goPath, m.env)
	if err != nil {
		m.Fail("Failed to build athens ", err)
	}
	stopAthens() // in case a dangling instance was around.
	// ignoring error as if no athens is running it fails.

	ctx := context.Background()
	ctx, m.stopAthens = context.WithCancel(ctx)
	runAthensAndWait(ctx, athensBin, m.getEnv())
	setupTestRepo(m.sampleRepoPath, "https://github.com/athens-artifacts/happy-path.git")
}

func (m *E2eSuite) TearDownSuite() {
	m.stopAthens()
	chmodR(m.goPath, 0o777)
	os.RemoveAll(m.goPath)
	chmodR(m.sampleRepoPath, 0o777)
	os.RemoveAll(m.sampleRepoPath)
}

func TestE2E(t *testing.T) {
	suite.Run(t, &E2eSuite{})
}

func (m *E2eSuite) SetupTest() {
	chmodR(m.goPath, 0o777)
	err := cleanGoCache(m.getEnv())
	if err != nil {
		m.Fail("Failed to clear go cache", err)
	}
}

func (m *E2eSuite) TestNoGoProxy() {
	cmd := exec.Command("go", "run", ".")
	cmd.Env = m.env
	cmd.Dir = m.sampleRepoPath

	err := cmd.Run()
	if err != nil {
		m.Fail("go run failed on test repo", err)
	}
}

func (m *E2eSuite) TestGoProxy() {
	cmd := exec.Command("go", "run", ".")
	cmd.Env = m.getEnvGoProxy(m.goPath)
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

func (m *E2eSuite) TestWrongGoProxy() {
	cmd := exec.Command("go", "run", ".")
	cmd.Env = m.getEnvWrongGoProxy(m.goPath)
	cmd.Dir = m.sampleRepoPath
	err := cmd.Run()
	m.Assert().NotNil(err, "Wrong proxy should fail")
}

func (m *E2eSuite) getEnv() []string {
	res := []string{
		fmt.Sprintf("GOPATH=%s", m.goPath),
		"GO111MODULE=on",
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("GOCACHE=%s", filepath.Join(m.goPath, "cache")),
	}
	return res
}

func (m *E2eSuite) getEnvGoProxy(gopath string) []string {
	res := m.getEnv()
	res = append(res, "GOPROXY=http://localhost:3000")
	return res
}

func (m *E2eSuite) getEnvWrongGoProxy(gopath string) []string {
	res := m.getEnv()
	res = append(res, "GOPROXY=http://localhost:3001")
	return res
}
