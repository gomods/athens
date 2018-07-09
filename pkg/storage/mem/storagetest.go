package mem

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
	memStore, err := NewStorage()

	return &StorageTest{
		storage: memStore,
		Model:   model,
	}, err
}

// TestNotFound tests whether storage returns ErrNotFound error on unknown package
func (st *StorageTest) TestNotFound() {
	_, err := st.storage.Get("some", "unknown")
	st.Require().Equal(true, storage.IsNotFoundError(err), "Invalid error type for %s: %#v", "In-Memory", err)
}
