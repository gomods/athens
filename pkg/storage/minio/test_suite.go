package minio

import (
	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/storage"
)

// TestSuite implements storage.TestSuite interface
type TestSuite struct {
	*suite.Model
	storage storage.Backend
}

// NewTestSuite creates a common test suite
func NewTestSuite(model *suite.Model) (storage.TestSuite, error) {
	endpoint := "127.0.0.1:9000"
	bucketName := "gomods"
	accessKeyID := "minio"
	secretAccessKey := "minio123"
	minioStorage, err := NewStorage(endpoint, accessKeyID, secretAccessKey, bucketName, false)

	return &TestSuite{
		storage: minioStorage,
		Model:   model,
	}, err
}

// Storage retrieves initialized storage backend
func (st *TestSuite) Storage() storage.Backend {
	return st.storage
}

// StorageHumanReadableName retrieves readable identifier of the storage
func (st *TestSuite) StorageHumanReadableName() string {
	return "Minio"
}
