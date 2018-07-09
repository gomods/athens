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

// TestNotFound tests whether storage returns ErrNotFound error on unknown package
func (st *TestSuite) TestNotFound() {
	_, err := st.storage.Get("some", "unknown")
	st.Require().Equal(true, storage.IsNotFoundError(err), "Invalid error type for %s: %#v", "FileSystem", err)
}
