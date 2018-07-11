package fs

import (
	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
)

// TestSuite implements storage.TestSuite interface
type TestSuite struct {
	*suite.Model
	storage storage.Backend
}

// NewTestSuite creates a common test suite
func NewTestSuite(model *suite.Model) (storage.TestSuite, error) {
	memFs := afero.NewOsFs()
	r, err := afero.TempDir(memFs, "", "athens-fs-storage-tests")
	if err != nil {
		return nil, err
	}

	fsStore, err := NewStorage(r, memFs)
	if err != nil {
		return nil, err
	}

	return &TestSuite{
		storage: fsStore,
		Model:   model,
	}, nil
}

// Storage retrieves initialized storage backend
func (st *TestSuite) Storage() storage.Backend {
	return st.storage
}

// StorageHumanReadableName retrieves readable identifier of the storage
func (st *TestSuite) StorageHumanReadableName() string {
	return "FileSystem"
}
