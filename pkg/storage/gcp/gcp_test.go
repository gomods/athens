package gcp

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/storage"
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
		err = s.delete(ctx, attrs.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func getStorage(t testing.TB) *Storage {
	url := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if url == "" {
		t.SkipNow()
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.GetTimeoutDuration(30))
	defer cancel()
	s, err := storage.NewClient(ctx)
	if err != nil {
		t.Fatalf("could not create new storage client: %s", err)
	}
	bucketName := "AthensTestBucket"
	bkt := s.Bucket(bucketName)
	if _, err := bkt.Attrs(ctx); err != nil {
		if err == storage.ErrBucketNotExist {
			t.Fatalf("bucket: %s does not exist - You must manually create a storage bucket for Athens, see https://cloud.google.com/storage/docs/creating-buckets#storage-create-bucket-console", bucketName)
		}
		t.Fatalf("error getting BucketHandle: %s", err)
	}

	return &Storage{
		bucket:       bkt,
		closeStorage: s.Close,
		timeout:      config.GetTimeoutDuration(300),
	}
}
