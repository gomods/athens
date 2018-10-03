package mem

import (
	"github.com/gomods/athens/pkg/storage"
)

// TestSuite implements storage.TestSuite interface
type TestSuite struct {
	storage storage.Backend
}

// NewTestSuite creates a common test suite
func NewTestSuite() (storage.TestSuite, error) {
	memStore, err := NewStorage()

	return &TestSuite{
		storage: memStore,
	}, err
}

// Storage retrieves initialized storage backend
func (ts *TestSuite) Storage() storage.Backend {
	return ts.storage
}

// StorageHumanReadableName retrieves readable identifier of the storage
func (ts *TestSuite) StorageHumanReadableName() string {
	return "In-memory"
}

// Cleanup tears down test
func (ts *TestSuite) Cleanup() error {
	return nil
}
