package module

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type CheckTests struct {
	suite.Suite
}

func TestCheck(t *testing.T) {
	t.Fatalf("TODO")
	suite.Run(t, new(DeleteTests))
}
