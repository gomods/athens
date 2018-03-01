package memory

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MemoryTests struct {
	suite.Suite
	mem GetterSaver
}

func (m *MemoryTests) BeforeTest() {
	m.mem = New()
}

func TestMemoryStorage(t *testing.T) {
	suite.Run(t, new(MemoryTests))
}
