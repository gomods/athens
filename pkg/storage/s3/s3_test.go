// +build !unit

package s3

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
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

func BenchmarkBackend(b *testing.B) {
	backend := getStorage(b)
	compliance.RunBenchmarks(b, backend, backend.clear)
}

func (s *Storage) clear() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	objects, err := s.s3API.ListObjectsWithContext(ctx, &s3.ListObjectsInput{Bucket: aws.String(s.bucket)})
	if err != nil {
		return err
	}

	for _, o := range objects.Contents {
		delParams := &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    o.Key,
		}

		_, err := s.s3API.DeleteObjectWithContext(ctx, delParams)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) createBucket() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	if _, err := s.s3API.CreateBucketWithContext(ctx, &s3.CreateBucketInput{Bucket: aws.String(s.bucket)}); err != nil {
		aerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}

		switch aerr.Code() {
		case s3.ErrCodeBucketAlreadyOwnedByYou:
			return nil
		case s3.ErrCodeBucketAlreadyExists:
			return nil
		default:
			return aerr
		}
	}

	return s.s3API.WaitUntilBucketExistsWithContext(ctx, &s3.HeadBucketInput{Bucket: aws.String(s.bucket)})
}

func getStorage(t testing.TB) *Storage {
	options := func(conf *aws.Config) {
		conf.Endpoint = aws.String(minioEndpoint)
		conf.DisableSSL = aws.Bool(true)
	}
	backend, err := New(
		&config.S3Config{
			Key:            "minio",
			Secret:         "minio123",
			Bucket:         "gomodsaws",
			Region:         "us-west-1",
			ForcePathStyle: true,
		},
		config.GetTimeoutDuration(300),
		options,
	)

	if err != nil {
		t.Fatal(err)
	}

	if err = backend.createBucket(); err != nil {
		t.Fatal(err)
	}

	return backend
}
