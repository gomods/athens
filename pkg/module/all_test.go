package module

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
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
