package minio

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/s3"
	minio "github.com/minio/minio-go/v6"
)

type storageImpl struct {
	storage.ToolchainStorage

	minioClient *minio.Client
	minioCore   *minio.Core
	bucketName  string
}

func (s *storageImpl) versionLocation(module, version string) string {
	return fmt.Sprintf("%s/%s", module, version)
}

// NewStorage returns a connected Minio or DigitalOcean Spaces storage
// that implements storage.Backend.
func NewStorage(conf *config.MinioConfig, timeout time.Duration) (storage.Backend, error) {
	const op errors.Op = "minio.NewStorage"
	endpoint := TrimHTTP(conf.Endpoint)
	accessKeyID := conf.Key
	secretAccessKey := conf.Secret
	bucketName := conf.Bucket
	region := conf.Region
	useSSL := conf.EnableSSL
	minioCore, err := minio.NewCore(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return nil, errors.E(op, err)
	}
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return nil, errors.E(op, err)
	}

	err = minioClient.MakeBucket(bucketName, region)
	if err != nil {
		// Check to see if we already own this bucket
		exists, err := minioClient.BucketExists(bucketName)
		if err != nil {
			return nil, errors.E(op, err)
		}
		if !exists {
			// MakeBucket Error takes priority
			return nil, errors.E(op, err)
		}
	}
	// The legacy MinIO client cannot conditionally publish objects. Reuse the
	// Athens S3 adapter with the same endpoint, credentials and bucket for the
	// release capability, so concurrent writers cannot replace stored archives.
	scheme := "http"
	if useSSL {
		scheme = "https"
	}
	if region == "" {
		region = "us-east-1"
	}
	client := awss3.NewFromConfig(aws.Config{Region: region, Credentials: credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")}, func(o *awss3.Options) { o.BaseEndpoint = aws.String(scheme + "://" + endpoint); o.UsePathStyle = true })
	releases := s3.NewWithClient(client, bucketName, timeout)
	return &storageImpl{minioClient: minioClient, minioCore: minioCore, bucketName: bucketName, ToolchainStorage: releases}, nil
}

// TrimHTTP trims "http://" or "https://" prefix from input string.
// Minio doesn't need to specify protocol in the URL so it should be trimmed.
// Related issue: https://github.com/gomods/athens/issues/1938#issuecomment-2067590653
func TrimHTTP(s string) string {
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "https://")
	return s
}
