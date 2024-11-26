package s3

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/gomods/athens/pkg/errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

func (s *Storage) clear() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	objects, err := s.s3API.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(s.bucket)})
	if err != nil {
		return err
	}

	for _, o := range objects.Contents {
		delParams := &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    o.Key,
		}

		_, err := s.s3API.DeleteObject(ctx, delParams)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) createBucket() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	if _, err := s.s3API.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(s.bucket)}); err != nil {
		var aerr smithy.APIError

		if errors.AsErr(err, &aerr) {
			switch aerr.(type) {
			case *types.BucketAlreadyOwnedByYou:
				return nil
			case *types.BucketAlreadyExists:
				return nil
			default:
				return aerr
			}
		}

		return err
	}

	waiter := s3.NewBucketExistsWaiter(s.s3API)

	return waiter.Wait(ctx, &s3.HeadBucketInput{Bucket: aws.String(s.bucket)}, 10*time.Minute)
}

func getStorage(t testing.TB) *Storage {
	url := os.Getenv("ATHENS_MINIO_ENDPOINT")
	if url == "" {
		t.SkipNow()
	}

	options := func(conf *aws.Config) {
		conf.BaseEndpoint = aws.String(url)
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
