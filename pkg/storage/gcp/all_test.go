package gcp

import (
	"context"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/option"
	"google.golang.org/appengine/aetest"
)

var (
	mod  = []byte{1, 2, 3}
	zip  = []byte{4, 5, 6}
	info = []byte{7, 8, 9}
)

type GcpTests struct {
	suite.Suite
	options option.ClientOption
	bucket  string
	module  string
	version string
}

func (g *GcpTests) SetupTest() {
	// this is the test project praxis-cab-207400.appspot.com
	// administered by robbie <robjloranger@protonmail.com>
	// the variable ATHENS_GCP_TEST_KEY should point to the test key json file
	// and must be set for this test which otherwise will be skipped
	creds := os.Getenv("ATHENS_GCP_TEST_KEY")
	if len(creds) == 0 {
		// TODO: skip/exit
	}
	g.options = option.WithCredentialsFile(creds)
	// time stamped test module names will prevent concurrent test interference
	g.bucket = "staging.praxis-cab-207400.appspot.com"
	g.module = "gcp-test" + time.Now().String()
	g.version = "v1.2.3"
}

func (g *GcpTests) TearDownTest() {
	ctx, done, err := aetest.NewContext()
	defer done()
	if err != nil {
		// TODO: don't panic
		panic(err)
	}
	client, err := storage.NewClient(ctx, g.options)
	if err != nil {
		panic(err)
	}

	bkt := client.Bucket(g.bucket)

	// remove all files and directories from this test
	err = cleanBucket(ctx, bkt, g.module, g.version)
	if err != nil {
		// TODO: don't panic
		panic(err)
	}
}

// cleanBucket iterates over the bucket contents and deletes everything
// matching the module name. folders do not exist so deleting the
// full object 'path' is sufficient
func cleanBucket(ctx context.Context, bkt *storage.BucketHandle, module, version string) error {
	return nil
}

func TestGcpStorage(t *testing.T) {
	suite.Run(t, new(GcpTests))
}
