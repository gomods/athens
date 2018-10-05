package s3

import (
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage"
	"github.com/minio/minio-go"
)

// TestSuite implements storage.TestSuite interface
type TestSuite struct {
	storage storage.Backend
	conf    *config.S3Config
}

// NewTestSuite creates a common test suite
func NewTestSuite(s3Conf *config.S3Config, cdnConf *config.CDNConfig) (storage.TestSuite, error) {
	s3Storage, err := New(s3Conf, cdnConf)

	return &TestSuite{
		storage: s3Storage,
		conf:    s3Conf,
	}, err
}

// Storage retrieves initialized storage backend
func (ts *TestSuite) Storage() storage.Backend {
	return ts.storage
}

// StorageHumanReadableName retrieves readable identifier of the storage
func (ts *TestSuite) StorageHumanReadableName() string {
	return "S3"
}

// Cleanup tears down test
func (ts *TestSuite) Cleanup() error {
	minioClient, _ := minio.New(ts.conf.Endpoint, ts.conf.Key, ts.conf.Secret, !ts.conf.DisableSSL)
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
