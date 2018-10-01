package fs

import (
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
)

// TestSuite implements storage.TestSuite interface
type TestSuite struct {
	storage storage.Backend
	fs      afero.Fs
	rootDir string
}

// NewTestSuite creates a common test suite
func NewTestSuite() (storage.TestSuite, error) {
	osFs := afero.NewOsFs()
	r, err := afero.TempDir(osFs, "", "athens-fs-storage-tests")
	if err != nil {
		return nil, err
	}

	fsStore, err := NewStorage(r, osFs)
	if err != nil {
		return nil, err
	}

	return &TestSuite{
		fs:      osFs,
		rootDir: r,
		storage: fsStore,
	}, nil
}

// Storage retrieves initialized storage backend
func (ts *TestSuite) Storage() storage.Backend {
	return ts.storage
}

// StorageHumanReadableName retrieves readable identifier of the storage
func (ts *TestSuite) StorageHumanReadableName() string {
	return "FileSystem"
}

// Cleanup tears down test
func (ts *TestSuite) Cleanup() error {
	return ts.fs.RemoveAll(ts.rootDir)
}
