package mongo

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/storage"
)

// TestSuite implements storage.TestSuite interface
type TestSuite struct {
	*suite.Model
	storage *ModuleStore
}

// NewTestSuite creates a common test suite
func NewTestSuite(model *suite.Model, mURI string, timeout time.Duration) (storage.TestSuite, error) {
	ms, err := newTestStore(mURI, timeout)
	if err != nil {
		return nil, err
	}
	return &TestSuite{
		storage: ms,
		Model:   model,
	}, err
}

func newTestStore(mURI string, timeout time.Duration) (*ModuleStore, error) {
	mongoStore, err := NewStorage(mURI, timeout)
	if err != nil {
		return nil, err
	}
	if mongoStore == nil {
		return nil, fmt.Errorf("Mongo storage is nil")
	}

	return mongoStore, nil
}

// Storage retrieves initialized storage backend
func (ts *TestSuite) Storage() storage.Backend {
	return ts.storage
}

// StorageHumanReadableName retrieves readable identifier of the storage
func (ts *TestSuite) StorageHumanReadableName() string {
	return "Mongo"
}

// Cleanup tears down test
func (ts *TestSuite) Cleanup() error {
	mURI := ts.storage.url
	s, err := mgo.DialWithTimeout(mURI, ts.storage.timeout)
	defer s.Close()
	if err != nil {
		return err
	}
	return s.DB("athens").C("modules").DropCollection()
}
