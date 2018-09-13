package minio

import (
	"fmt"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	minio "github.com/minio/minio-go"
)

type storageImpl struct {
	minioClient *minio.Client
	bucketName  string
}

func (s *storageImpl) versionLocation(module, version string) string {
	return fmt.Sprintf("%s/%s", module, version)
}

// NewStorage returns a new ListerSaver implementation that stores
// everything under rootDir
func NewStorage(conf *config.MinioConfig) (storage.Backend, error) {
	const op errors.Op = "minio.NewStorage"
	endpoint := conf.Endpoint
	accessKeyID := conf.Key
	secretAccessKey := conf.Secret
	bucketName := conf.Bucket
	useSSL := conf.EnableSSL
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return nil, errors.E(op, err)
	}

	err = minioClient.MakeBucket(bucketName, "")
	if err != nil {
		// Check to see if we already own this bucket
		exists, err := minioClient.BucketExists(bucketName)
		if err == nil && !exists {
			return nil, errors.E(op, err)
		}
	}
	return &storageImpl{minioClient, bucketName}, nil
}
