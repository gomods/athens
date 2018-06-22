package gcp

import (
	"context"
	"fmt"
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
	context context.Context
	options option.ClientOption
	bucket  string
	module  string
	version string
}

func (g *GcpTests) SetupSuite() {
	// this is the test project praxis-cab-207400.appspot.com
	// administered by robbie <robjloranger@protonmail.com>
	// the variable ATHENS_GCP_TEST_KEY should point to the test key json file
	// and must be set for this test which otherwise will be skipped
	creds := os.Getenv("ATHENS_GCP_TEST_KEY")
	if len(creds) == 0 {
		g.T().Skip()
	}
	g.options = option.WithCredentialsFile(creds)
	ctx, done, err := aetest.NewContext()
	defer done()
	if err != nil {
		// TODO: don't panic
		panic(err)
	}
	g.context = ctx
	// time stamped test module names will prevent concurrent test interference
	g.bucket = "staging.praxis-cab-207400.appspot.com"
	g.module = "gcp-test" + time.Now().String()
	g.version = "v1.2.3"
}

func (g *GcpTests) TearDownSuite() {
	client, err := storage.NewClient(g.context, g.options)
	if err != nil {
		panic(err)
	}

	bkt := client.Bucket(g.bucket)

	// remove all files and directories from this test
	err = cleanBucket(g.context, bkt, g.module, g.version)
	if err != nil {
		// TODO: don't panic
		panic(err)
	}
}

// cleanBucket iterates over the bucket contents and deletes everything
// matching the module name. folders do not exist so deleting the
// full object 'path' is sufficient
func cleanBucket(ctx context.Context, bkt *storage.BucketHandle, mod, ver string) error {
	if err := bkt.Object(fmt.Sprintf("%s/@v/%s.%s", mod, ver, "mod")).Delete(ctx); err != nil {
		return err
	}
	if err := bkt.Object(fmt.Sprintf("%s/@v/%s.%s", mod, ver, "info")).Delete(ctx); err != nil {
		return err
	}
	if err := bkt.Object(fmt.Sprintf("%s/@v/%s.%s", mod, ver, "zip")).Delete(ctx); err != nil {
		return err
	}
	return nil
}

func TestGcpStorage(t *testing.T) {
	suite.Run(t, new(GcpTests))
}
