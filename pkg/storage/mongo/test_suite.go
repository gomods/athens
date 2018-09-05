package mongo

import (
	"fmt"
	"path/filepath"

	"github.com/globalsign/mgo"
	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage"
)

var (
	testConfigFile = filepath.Join("../../../config.test.toml")
)

// TestSuite implements storage.TestSuite interface
type TestSuite struct {
	*suite.Model
	storage *ModuleStore
}

// NewTestSuite creates a common test suite
func NewTestSuite(model *suite.Model, configFile string) (storage.TestSuite, error) {
	ms, err := newTestStore(configFile)
	if err != nil {
		return nil, err
	}
	return &TestSuite{
		storage: ms,
		Model:   model,
	}, err
}

func newTestStore(configFile string) (*ModuleStore, error) {
	conf, err := config.GetConf(configFile)
	if err != nil {
		return nil, err
	}
	mongoStore, err := NewStorage(conf.Storage.Mongo)
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
	conf, err := config.GetConf(testConfigFile)
	if err != nil {
		return err
	}
	if conf.Storage == nil || conf.Storage.Mongo == nil {
		return fmt.Errorf("Invalid Mongo Storage Provided")
	}
	s, err := mgo.DialWithTimeout(conf.Storage.Mongo.URL, conf.Storage.Mongo.TimeoutDuration())
	defer s.Close()
	if err != nil {
		return err
	}
	return s.DB("athens").C("modules").DropCollection()
}
