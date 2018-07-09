package mongo

import (
	"fmt"

	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/config/env"
	"github.com/gomods/athens/pkg/storage"
)

// StorageTest implements StorageTest interface
type StorageTest struct {
	*suite.Model
	storage storage.Backend
}

// NewStorageTest creates a common test suite
func NewStorageTest(model *suite.Model) (storage.StorageTest, error) {
	muri, err := env.MongoURI()
	if err != nil {
		return nil, err
	}

	mongoStore := NewStorage(muri)
	if mongoStore == nil {
		return nil, fmt.Errorf("Mongo storage is nil")
	}

	err = mongoStore.Connect()

	return &StorageTest{
		storage: mongoStore,
		Model:   model,
	}, err
}

// TestNotFound tests whether storage returns ErrNotFound error on unknown package
func (st *StorageTest) TestNotFound() {
	_, err := st.storage.Get("some", "unknown")
	st.Require().Equal(true, storage.IsNotFoundError(err), "Invalid error type for %s: %#v", "Minio", err)
}
