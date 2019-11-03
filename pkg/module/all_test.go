package module

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/gobuffalo/envy"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

const (
	// these values need to point to a real repository that has a tag
	// github.com/NYTimes/gizmo is a example of a path that needs to be encoded so we can cover that case as well
	repoURI = "github.com/NYTimes/gizmo"
	version = "v0.1.4"
)

type cmdFunc func(s ...string) *exec.Cmd

type ModuleSuite struct {
	suite.Suite
	fs           afero.Fs
	goBinaryName string
	env          []string
	realCmdFunc  cmdFunc
}

func (s *ModuleSuite) SetupTest() {
	s.fs = afero.NewMemMapFs()
	s.realCmdFunc = cmd
	cmd = helperProcess
}

func (s *ModuleSuite) AfterTest(suiteName, testName string) {
	cmd = s.realCmdFunc
}

func (s *ModuleSuite) useRealGoBin() {
	cmd = s.realCmdFunc
}

func TestModules(t *testing.T) {
	goBinaryPath := envy.Get("GO_BINARY_PATH", "go")
	suite.Run(t, &ModuleSuite{goBinaryName: goBinaryPath, env: []string{"GOPROXY=direct"}})
}

// helper function that redirects the execution to a mock process implemented by
// TestHelperProcess
func helperProcess(s ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, s...)
	env := []string{
		"GO_WANT_HELPER_PROCESS=1",
	}

	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(env, os.Environ()...)
	return cmd
}

// This is not a real test. It's used in case helperProcess is used to build exec.Cmd.
// In that case the command execution is redirected to this function.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		return
	}

	args := os.Args

	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	switch {
	case fmt.Sprint(args) == "[go mod download -json github.com/NYTimes/gizmo@v0.1.4]":
		fmt.Println(`{
			"Path": "github.com/NYTimes/gizmo",
			"Version": "v0.1.4",
			"Info": "test_data/gizmo/@v/v0.1.4.info",
			"GoMod": "test_data/gizmo/@v/v0.1.4.mod",
			"Zip": "test_data/gizmo/@v/v0.1.4.zip",
			"Dir": "/tmp/athens874423942/pkg/mod/github.com/!n!y!times/gizmo@v0.1.4",
			"Sum": "h1:ordP01G3DLAPuNk5+sa5hha8sBZPGp9SzNRdM/Sj2Jw=",
			"GoModSum": "h1:JfXBn0N9slseaeNYjK/B+/AoKaUgR1wBuw71IGAG1fw="
	}`)
		os.Exit(0)
	case fmt.Sprint(args) == "[go mod download -json laks47dfjoijskdvjxuyyd.com/pkg/errors@v0.8.1]":
		fmt.Println("{}")
		os.Exit(1)
	}
}
