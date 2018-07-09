package mongo

import (
	"fmt"

	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/config/env"
	"github.com/gomods/athens/pkg/storage"
)

// TestSuite implements storage.TestSuite interface
type TestSuite struct {
	*suite.Model
	storage storage.Backend
}

// NewTestSuite creates a common test suite
func NewTestSuite(model *suite.Model) (storage.TestSuite, error) {
	muri, err := env.MongoURI()
	if err != nil {
		return nil, err
	}

	mongoStore := NewStorage(muri)
	if mongoStore == nil {
		return nil, fmt.Errorf("Mongo storage is nil")
	}

	err = mongoStore.Connect()

	return &TestSuite{
		storage: mongoStore,
		Model:   model,
	}, err
}

// TestNotFound tests whether storage returns ErrNotFound error on unknown package
func (st *TestSuite) TestNotFound() {
	_, err := st.storage.Get("some", "unknown")
	st.Require().Equal(true, storage.IsNotFoundError(err), "Invalid error type for %s: %#v", "Minio", err)
}
