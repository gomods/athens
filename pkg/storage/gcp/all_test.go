package gcp

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/stretchr/testify/suite"
)

var (
	mod  = []byte{1, 2, 3}
	zip  = []byte{4, 5, 6}
	info = []byte{7, 8, 9}
)

const (
	testConfigFile = "../../../config.test.toml"
)

type GcpTests struct {
	suite.Suite
	context context.Context
	module  string
	version string
	store   *Storage
	url     *url.URL
	bucket  *bucketMock
	timeout time.Duration
}

func (g *GcpTests) SetupSuite() {
	g.context = context.Background()
	g.module = "gcp-test"
	g.version = "v1.2.3"
	g.url, _ = url.Parse("https://storage.googleapis.com/testbucket")
	g.bucket = newBucketMock()
	g.store = newWithBucket(g.bucket, g.url, g.timeout)
}

func TestGcpStorage(t *testing.T) {
	conf := config.GetConfLogErr(testConfigFile, t)
	if conf.Storage == nil || conf.Storage.GCP == nil {
		t.Fatalf("Invalid GCP config provided")
	}
	gcpTimeout := config.TimeoutDuration(conf.Storage.GCP.Timeout)
	suite.Run(t, &GcpTests{
		timeout: gcpTimeout,
	})
}
