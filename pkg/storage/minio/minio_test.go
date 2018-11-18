package minio

import (
	"os"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage/compliance"
)

func TestBackend(t *testing.T) {
	backend := getStorage(t)
	compliance.RunTests(t, backend, backend.clear)
}

func BenchmarkBackend(b *testing.B) {
	backend := getStorage(b)
	compliance.RunBenchmarks(b, backend, backend.clear)
}

func (s *storageImpl) clear() error {
	doneCh := make(chan struct{})
	defer close(doneCh)
	objectCh := s.minioClient.ListObjectsV2(s.bucketName, "", true, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			return object.Err
		}
		if err := s.minioClient.RemoveObject(s.bucketName, object.Key); err != nil {
			return err
		}
	}
	return nil
}

func getStorage(t testing.TB) *storageImpl {
	url := os.Getenv("ATHENS_MINIO_ENDPOINT")
	if url == "" {
		t.SkipNow()
	}

	backend, err := NewStorage(&config.MinioConfig{
		Endpoint: url,
		Key:      "minio",
		Secret:   "minio123",
		Bucket:   "gomods",
	}, config.GetTimeoutDuration(300))
	if err != nil {
		t.Fatal(err)
	}

	return backend.(*storageImpl)
}
