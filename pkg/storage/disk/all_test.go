package disk

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type DiskTests struct {
	suite.Suite
}

func TestDiskStorage(t *testing.T) {
	suite.Run(t, new(DiskTests))
}
