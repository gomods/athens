package minio

import (
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage"
	minio "github.com/minio/minio-go"
	"github.com/stretchr/testify/suite"
)

type MinioTests struct {
	suite.Suite
	storage                                            storage.Backend
	endpoint, accessKeyID, secretAccessKey, bucketName string
}

func (d *MinioTests) SetupTest() {
	// TODO: what is the difference between all_test and test_suite.go??
	d.endpoint = "127.0.0.1:9000"
	d.bucketName = "gomods"
	d.accessKeyID = "minio"
	d.secretAccessKey = "minio123"
	conf := &config.MinioConfig{
		Endpoint:  d.endpoint,
		Bucket:    d.bucketName,
		Key:       d.accessKeyID,
		Secret:    d.secretAccessKey,
		EnableSSL: false,
	}
	storage, err := NewStorage(conf)
	d.Require().NoError(err)
	d.storage = storage
}

func (d *MinioTests) TearDownTest() {
	minioClient, _ := minio.New(d.endpoint, d.accessKeyID, d.secretAccessKey, false)
	doneCh := make(chan struct{})
	defer close(doneCh)
	objectCh := minioClient.ListObjectsV2(d.bucketName, "", true, doneCh)
	for object := range objectCh {
		d.Require().NoError(minioClient.RemoveObject(d.bucketName, object.Key))
	}
}

func TestMinioStorage(t *testing.T) {
	suite.Run(t, new(MinioTests))
}
