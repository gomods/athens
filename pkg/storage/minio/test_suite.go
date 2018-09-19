package minio

import (
	"fmt"

	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage"
	minio "github.com/minio/minio-go"
)

// TestSuite implements storage.TestSuite interface
type TestSuite struct {
	*suite.Model
	storage storage.Backend
	conf    *config.MinioConfig
}

// NewTestSuite creates a common test suite
func NewTestSuite(model *suite.Model, configFile string) (storage.TestSuite, error) {
	conf, err := config.GetConf(configFile)
	if err != nil {
		return nil, err
	}
	minioStorage, err := newTestStore(conf.Storage.Minio)

	return &TestSuite{
		storage: minioStorage,
		Model:   model,
		conf:    conf.Storage.Minio,
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
