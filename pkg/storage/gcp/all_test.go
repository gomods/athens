package gcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

var (
	mod  = []byte{1, 2, 3}
	zip  = []byte{4, 5, 6}
	info = []byte{7, 8, 9}
)

type GcpTests struct {
	suite.Suite
	context context.Context
	module  string
	version string
	store   *Storage
}

func (g *GcpTests) SetupSuite() {
	g.context = context.Background()
	g.module = "gcp-test"
	g.version = "v1.2.3"
	g.store = newWithBucket(newBucketMock())
}

func TestGcpStorage(t *testing.T) {
	suite.Run(t, new(GcpTests))
}
