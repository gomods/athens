package gcp

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage/compliance"
	"github.com/technosophos/moniker"
	"google.golang.org/api/iterator"
)

func TestBackend(t *testing.T) {
	backend := getStorage(t)
	compliance.RunTests(t, backend, backend.clear)
}

func BenchmarkBackend(b *testing.B) {
	backend := getStorage(b)
	compliance.RunBenchmarks(b, backend, backend.clear)
}

func (s *Storage) clear() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	it := s.bucket.Objects(ctx, nil)

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		err = s.bucket.Object(attrs.Name).Delete(ctx)
		if err != nil {
			return err
		}
	}
	// since this is a unique bucket name (see getTestConfig), we need to
	// not only delete all the objects in the bucket (above), we need to then
	// delete the bucket itself
	return s.bucket.Delete(ctx)
}

func getStorage(t testing.TB) *Storage {
	t.Helper()
	cfg := getTestConfig()
	if cfg == nil {
		t.SkipNow()
	}
	ctx := context.Background()

	// get the raw GCP client
	cl, err := newClient(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Create the GCP bucket, so that the below call to New doesn't fail.
	//
	// The 'New' function calls bucket.Attrs(), so the bucket has to exist
	bkt := cl.Bucket(cfg.Bucket)
	if err := bkt.Create(ctx, cfg.ProjectID, nil); err != nil {
		t.Fatal(err)
	}

	// create a new Storage implementation for GCP
	s, err := New(ctx, cfg, config.GetTimeoutDuration(30))
	if err != nil {
		t.Fatal(err)
	}

	return s
}

func getTestConfig() *config.GCPConfig {
	// moniker is a cool library to produce mostly unique, human-readable names
	// see https://github.com/technosophos/moniker for more details
	namer := moniker.New()
	bucketName := fmt.Sprintf("athens_drone_%s", namer.NameSep("_"))
	creds := os.Getenv("GCS_SERVICE_ACCOUNT")
	if creds == "" {
		return nil
	}
	projID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projID == "" {
		return nil
	}
	return &config.GCPConfig{
		Bucket:    bucketName,
		JSONKey:   creds,
		ProjectID: projID,
	}
}
