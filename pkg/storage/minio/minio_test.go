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

// TestNewStorageExists tests the logic around MakeBucket and BucketExists
func TestNewStorageExists(t *testing.T) {
	url := TrimHTTP(os.Getenv("ATHENS_MINIO_ENDPOINT"))
	if url == "" {
		t.SkipNow()
	}

	tests := []struct {
		name         string
		deleteBucket bool
	}{
		{"testbucket", false}, // test creation
		{"testbucket", true},  // test exists
	}

	for _, test := range tests {
		backend, err := NewStorage(&config.MinioConfig{
			Endpoint: url,
			Key:      "minio",
			Secret:   "minio123",
			Bucket:   test.name,
		}, config.GetTimeoutDuration(300))
		if err != nil {
			t.Fatalf("TestNewStorageExists failed for bucketname:  %s, error: %v\n", test.name, err)
		}

		client, ok := backend.(*storageImpl)
		if test.deleteBucket && ok {
			client.minioClient.RemoveBucket(test.name)
		}
	}
}

// TestNewStorageError tests the logic around MakeBucket and BucketExists
// MakeBucket uses a strict naming path in minio while BucketExists does not.
// To ensure both paths are tested, there is a strict path error using the
// "_" and a non strict error using less than 3 characters
func TestNewStorageError(t *testing.T) {
	url := TrimHTTP(os.Getenv("ATHENS_MINIO_ENDPOINT"))
	if url == "" {
		t.SkipNow()
	}

	// "_" is not allowed in a bucket name
	// bucket name must be bigger than 3
	tests := []string{"test_bucket", "1"}

	for _, bucketName := range tests {
		_, err := NewStorage(&config.MinioConfig{
			Endpoint: url,
			Key:      "minio",
			Secret:   "minio123",
			Bucket:   bucketName,
		}, config.GetTimeoutDuration(300))
		if err == nil {
			t.Fatalf("TestNewStorageError failed for bucketname:  %s\n", bucketName)
		}
	}
}

func BenchmarkBackend(b *testing.B) {
	backend := getStorage(b)
	compliance.RunBenchmarks(b, backend, backend.clear)
}

func (s *storageImpl) clear() error {
	objectCh, _ := s.minioCore.ListObjectsV2(s.bucketName, "", "", false, "", 0, "")
	for _, object := range objectCh.Contents {
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
	url := TrimHTTP(os.Getenv("ATHENS_MINIO_ENDPOINT"))
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
