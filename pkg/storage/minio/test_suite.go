package minio

import (
	"fmt"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage"
	minio "github.com/minio/minio-go"
)

// TestSuite implements storage.TestSuite interface
type TestSuite struct {
	storage storage.Backend
	conf    *config.MinioConfig
}

// NewTestSuite creates a common test suite
func NewTestSuite(conf *config.MinioConfig) (storage.TestSuite, error) {
	minioStorage, err := newTestStore(conf)

	return &TestSuite{
		storage: minioStorage,
		conf:    conf,
	}, err
}

func newTestStore(conf *config.MinioConfig) (storage.Backend, error) {
	minioStore, err := NewStorage(conf)
	if err != nil {
		return nil, fmt.Errorf("Not able to connect to minio storage: %s", err.Error())
	}

	return minioStore, nil
}

// Storage retrieves initialized storage backend
func (ts *TestSuite) Storage() storage.Backend {
	return ts.storage
}

// StorageHumanReadableName retrieves readable identifier of the storage
func (ts *TestSuite) StorageHumanReadableName() string {
	return "Minio"
}

// Cleanup tears down test
func (ts *TestSuite) Cleanup() error {
	minioClient, _ := minio.New(ts.conf.Endpoint, ts.conf.Key, ts.conf.Secret, ts.conf.EnableSSL)
	doneCh := make(chan struct{})
	defer close(doneCh)
	objectCh := minioClient.ListObjectsV2(ts.conf.Bucket, "", true, doneCh)
	for object := range objectCh {
		//TODO: could return multi error and clean other objects
		if err := minioClient.RemoveObject(ts.conf.Bucket, object.Key); err != nil {
			return err
		}
	}
	return nil
}
