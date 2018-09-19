package mongo

import (
	"fmt"

	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage"
)

// TestSuite implements storage.TestSuite interface
type TestSuite struct {
	*suite.Model
	storage *ModuleStore
	conf    *config.MongoConfig
}

// NewTestSuite creates a common test suite
func NewTestSuite(model *suite.Model, conf *config.MongoConfig) (storage.TestSuite, error) {
	ms, err := newTestStore(conf)
	if err != nil {
		return nil, err
	}
	return &TestSuite{
		storage: ms,
		Model:   model,
		conf:    conf,
	}, err
}

func newTestStore(conf *config.MongoConfig) (*ModuleStore, error) {
	mongoStore, err := NewStorage(conf)
	if err != nil {
		return nil, fmt.Errorf("Not able to connect to mongo storage: %s", err.Error())
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
	ms, err := newTestStore(ts.conf)
	if err != nil {
		return err
	}
	_, err = ms.s.DB(ms.d).C(ms.c).RemoveAll(nil)
	return err
}
