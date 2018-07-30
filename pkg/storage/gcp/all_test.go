package gcp

import (
	"context"
	"fmt"
	"net/url"
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
	url     *url.URL
}

func (g *GcpTests) SetupSuite() {
	g.context = context.Background()
	g.module = "gcp-test"
	g.version = "v1.2.3"
	g.url, _ = url.Parse(fmt.Sprintf("https://storage.googleapis.com/testbucket"))

	g.store = newWithBucket(newBucketMock(), g.url)
}

func TestGcpStorage(t *testing.T) {
	suite.Run(t, new(GcpTests))
}
