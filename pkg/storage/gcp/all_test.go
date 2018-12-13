package gcp

import (
	"context"
	"flag"
	"net/url"
	"testing"
	"time"

	"github.com/gobuffalo/envy"
	"github.com/gomods/athens/pkg/config"
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
	bucket  *bucketMock
}

var realGcp = flag.Bool("realgcp", false, "tests against a real gcp instance")
var project = flag.String("gcpprj", "", "the gcp project to test against")
var bucket = flag.String("gcpbucket", "", "the gcp bucket to test against")

func (g *GcpTests) SetupSuite() {
	g.context = context.Background()
	g.module = "github.com/foo/gcp-test" + time.Now().String()
	g.version = "v1.2.3"

	if !*realGcp {
		setupMockStorage(g)
	} else {
		setupRealStorage(g)
	}
}

func TestGcpStorage(t *testing.T) {
	suite.Run(t, new(GcpTests))
}

func (g *GcpTests) BucketReadClosed() bool {
	if g.bucket != nil {
		return g.bucket.ReadClosed()
	}
	return true
}

func (g *GcpTests) BucketWriteClosed() bool {
	if g.bucket != nil {
		return g.bucket.WriteClosed()
	}
	return true
}

func setupMockStorage(g *GcpTests) {
	g.url, _ = url.Parse("https://storage.googleapis.com/testbucket")
	g.bucket = newBucketMock()
	g.store = newWithBucket(g.bucket, g.url, time.Second)
}

func setupRealStorage(g *GcpTests) {
	_, err := envy.MustGet("GOOGLE_APPLICATION_CREDENTIALS")
	if err != nil {
		g.T().Skip()
	}
	if *project == "" || *bucket == "" {
		g.T().Skip()
	}

	g.store, err = New(context.Background(), &config.GCPConfig{
		ProjectID: *project,
		Bucket:    *bucket,
	}, 300*time.Second)
	g.Require().NoError(err)
}
