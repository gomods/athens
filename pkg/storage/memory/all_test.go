package memory

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MemoryTests struct {
	suite.Suite
}

func TestMemoryStorage(t *testing.T) {
	suite.Run(t, new(MemoryTests))
}
