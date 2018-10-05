package s3

import (
	"os"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage/compliance"
	"github.com/minio/minio-go"
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
	minioClient, _ := minio.New(os.Getenv("ATHENS_MINIO_ENDPOINT"), "minio", "minio123", false)
	doneCh := make(chan struct{})
	defer close(doneCh)
	objectCh := minioClient.ListObjectsV2("gomods", "", true, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			return object.Err
		}
		if err := minioClient.RemoveObject("gomods", object.Key); err != nil {
			return err
		}
	}
	return nil
}

func getStorage(t testing.TB) *Storage {
	url := os.Getenv("ATHENS_MINIO_ENDPOINT")
	if url == "" {
		t.SkipNow()
	}

	backend, err := New(
		&config.S3Config{
			Endpoint:       url,
			Key:            "minio",
			Secret:         "minio123",
			Bucket:         "gomods",
			Region:         "us-west-1",
			DisableSSL:     true,
			ForcePathStyle: true,
		},
		&config.CDNConfig{
			Endpoint: "cdn.example.com",
			TimeoutConf: config.TimeoutConf{
				Timeout: 300,
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return backend
}
