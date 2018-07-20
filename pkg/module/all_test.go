package module

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

const (
	// these values need to point to a real repository that has a tag
	repoURI = "github.com/arschles/assert"
	version = "v1.0.0"
)

type ModuleSuite struct {
	suite.Suite
	fs afero.Fs
}

func (m *ModuleSuite) SetupTest() {
	m.fs = afero.NewMemMapFs()
}

func TestModules(t *testing.T) {
	suite.Run(t, &ModuleSuite{})
}
