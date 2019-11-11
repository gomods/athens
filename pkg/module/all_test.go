package module

import (
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

type ModuleSuite struct {
	suite.Suite
	fs           afero.Fs
	goBinaryName string
	env          []string
}

func (m *ModuleSuite) SetupTest() {
	m.fs = afero.NewMemMapFs()
}

func TestModules(t *testing.T) {
	goBinaryPath := envy.Get("GO_BINARY_PATH", "go")
	suite.Run(t, &ModuleSuite{goBinaryName: goBinaryPath, env: []string{"GOPROXY=direct"}})
}
