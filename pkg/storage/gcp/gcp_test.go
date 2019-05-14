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
	defer backend.bucket.Delete(context.Background())
	compliance.RunTests(t, backend, backend.clear)
}

func BenchmarkBackend(b *testing.B) {
	backend := getStorage(b)
	defer backend.bucket.Delete(context.Background())
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
	return nil
}

func getStorage(t testing.TB) *Storage {
	t.Helper()
	bucketName := randomBucketName(os.Getenv("DRONE_PULL_REQUEST"))
	cfg := getTestConfig(bucketName)
	if cfg == nil {
		t.SkipNow()
	}

	s, err := newClient(context.Background(), cfg, config.GetTimeoutDuration(30))
	if err != nil {
		t.Fatal(err)
	}
	err = s.bucket.Create(context.Background(), cfg.ProjectID, nil)
	if err != nil {
		t.Fatal(err)
	}

	return s
}

func getTestConfig(bucket string) *config.GCPConfig {
	creds := os.Getenv("GCS_SERVICE_ACCOUNT")
	if creds == "" {
		return nil
	}
	return &config.GCPConfig{
		Bucket:    bucket,
		JSONKey:   creds,
		ProjectID: os.Getenv("GCS_PROJECT_ID"),
	}
}

func randomBucketName(prefix string) string {
	// moniker is a cool library to produce mostly unique, human-readable names
	// see https://github.com/technosophos/moniker for more details
	namer := moniker.New()
	if prefix != "" {
		return fmt.Sprintf("%s_%s", prefix, namer.NameSep("_"))
	}
	return namer.NameSep("_")
}
