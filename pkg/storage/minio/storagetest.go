package minio

import (
	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/storage"
)

// StorageTest implements StorageTest interface
type StorageTest struct {
	*suite.Model
	storage storage.Backend
}

// NewStorageTest creates a common test suite
func NewStorageTest(model *suite.Model) (storage.StorageTest, error) {
	endpoint := "127.0.0.1:9000"
	bucketName := "gomods"
	accessKeyID := "minio"
	secretAccessKey := "minio123"
	minioStorage, err := NewStorage(endpoint, accessKeyID, secretAccessKey, bucketName, false)

	return &StorageTest{
		storage: minioStorage,
		Model:   model,
	}, err
}

// TestNotFound tests whether storage returns ErrNotFound error on unknown package
func (st *StorageTest) TestNotFound() {
	_, err := st.storage.Get("some", "unknown")
	st.Require().Equal(true, storage.IsNotFoundError(err), "Invalid error type for %s: %#v", "Minio", err)
}
