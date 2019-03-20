package gcp

import (
	"context"
	"os"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage/compliance"
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
	return nil
}

func getStorage(t testing.TB) *Storage {
	t.Helper()
	cfg := getTestConfig()
	if cfg == nil {
		t.SkipNow()
	}

	s, err := New(context.Background(), cfg, config.GetTimeoutDuration(30))
	if err != nil {
		t.Fatal(err)
	}

	return s
}

func getTestConfig() *config.GCPConfig {
	creds := os.Getenv("GCS_SERVICE_ACCOUNT")
	if creds == "" {
		return nil
	}
	return &config.GCPConfig{
		Bucket:  "athens_drone_bucket",
		JSONKey: creds,
	}
}
