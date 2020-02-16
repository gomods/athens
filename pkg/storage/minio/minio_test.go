// +build !unit

package minio

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage/compliance"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var useTestContainers = os.Getenv("ATHENS_USE_TEST_CONTAINERS")
var minioEndpoint string

func TestMain(m *testing.M) {
	if useTestContainers != "1" {
		os.Exit(m.Run())
		minioEndpoint = os.Getenv("ATHENS_MINIO_ENDPOINT")
	}

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "minio/minio:latest",
		ExposedPorts: []string{"9000/tcp"},
		WaitingFor:   wait.ForLog("Endpoint:").WithStartupTimeout(time.Minute * 1),
		Cmd:          []string{"server", "/data"},
		Env: map[string]string{
			"MINIO_ACCESS_KEY": "minio",
			"MINIO_SECRET_KEY": "minio123",
		},
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		panic(err.Error())
	}

	ep, err := c.Endpoint(context.Background(), "")
	if err != nil {
		panic(err.Error())
	}

	minioEndpoint = ep
	defer c.Terminate(ctx)
	os.Exit(m.Run())

}

func TestBackend(t *testing.T) {
	backend := getStorage(t)
	compliance.RunTests(t, backend, backend.clear)
}

// TestNewStorageExists tests the logic around MakeBucket and BucketExists
func TestNewStorageExists(t *testing.T) {

	tests := []struct {
		name         string
		deleteBucket bool
	}{
		{"testbucket", false}, // test creation
		{"testbucket", true},  // test exists
	}

	for _, test := range tests {
		backend, err := NewStorage(&config.MinioConfig{
			Endpoint: minioEndpoint,
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

	// "_" is not allowed in a bucket name
	// bucket name must be bigger than 3
	tests := []string{"test_bucket", "1"}

	for _, bucketName := range tests {
		_, err := NewStorage(&config.MinioConfig{
			Endpoint: minioEndpoint,
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
	backend, err := NewStorage(&config.MinioConfig{
		Endpoint: minioEndpoint,
		Key:      "minio",
		Secret:   "minio123",
		Bucket:   "gomods",
	}, config.GetTimeoutDuration(300))
	if err != nil {
		t.Fatal(err)
	}

	return backend.(*storageImpl)
}
