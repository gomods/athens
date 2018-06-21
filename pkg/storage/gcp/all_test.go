package gcp

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/api/option"
)

var (
	mod  = []byte{1, 2, 3}
	zip  = []byte{4, 5, 6}
	info = []byte{7, 8, 9}
)

type GcpTests struct {
	suite.Suite
	options option.ClientOption
}

func (g *GcpTests) SetupTest() {
	// this is the test project praxis-cab-207400.appspot.com
	// administered by robbie <robjloranger@protonmail.com>
	// the test keys provide bucket controls only
	// please use staging.praxis-cab-207400.appspot.com
	g.options = option.WithCredentialsFile("test-keys.json")
}

func TestGcpStorage(t *testing.T) {
	suite.Run(t, new(GcpTests))
}
