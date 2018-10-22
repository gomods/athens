package s3

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
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
	objects, err := s.s3API.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(s.bucket)})
	if err != nil {
		return err
	}

	for _, o := range objects.Contents {
		delParams := &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    o.Key,
		}

		_, err := s.s3API.DeleteObject(delParams)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) createBucket() error {
	if _, err := s.s3API.CreateBucket(&s3.CreateBucketInput{Bucket: aws.String(s.bucket)}); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				return nil
			case s3.ErrCodeBucketAlreadyExists:
				return nil
			default:
				return aerr
			}
		} else {
			return err
		}
	}

	if err := s.s3API.WaitUntilBucketExists(&s3.HeadBucketInput{Bucket: aws.String(s.bucket)}); err != nil {
		return err
	}

	return nil
}

func getStorage(t testing.TB) *Storage {
	options := func(conf *aws.Config) {
		conf.Endpoint = aws.String("127.0.0.1:9001")
		conf.DisableSSL = aws.Bool(true)
		conf.S3ForcePathStyle = aws.Bool(true)
	}
	backend, err := New(
		&config.S3Config{
			Key:    "minio",
			Secret: "minio123",
			Bucket: "gomodsaws",
			Region: "us-west-1",
			TimeoutConf: config.TimeoutConf{
				Timeout: 300,
			},
		},
		nil,
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
