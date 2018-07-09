package mem

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
	memStore, err := NewStorage()

	return &TestSuite{
		storage: memStore,
		Model:   model,
	}, err
}

// TestNotFound tests whether storage returns ErrNotFound error on unknown package
func (st *TestSuite) TestNotFound() {
	_, err := st.storage.Get("some", "unknown")
	st.Require().Equal(true, storage.IsNotFoundError(err), "Invalid error type for %s: %#v", "In-Memory", err)
}
