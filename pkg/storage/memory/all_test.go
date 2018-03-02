package memory

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	baseURL = "base.com"
	module  = "my/module"
)

type MemoryTests struct {
	suite.Suite
	mem GetterSaver
}

func (m *MemoryTests) SetupTest() {
	m.mem = New()
}

func TestMemoryStorage(t *testing.T) {
	suite.Run(t, new(MemoryTests))
}
