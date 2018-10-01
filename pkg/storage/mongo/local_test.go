package mongo

import (
	"path/filepath"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/stretchr/testify/suite"
)

var (
	testConfigFile = filepath.Join("..", "..", "..", "config.dev.toml")
)

type MongoTests struct {
	suite.Suite
}

func (m *MongoTests) SetupTest() {
	conf, err := config.GetConf(testConfigFile)
	if err != nil {
		m.T().Fatalf("Unable to parse config file: %s", err.Error())
	}

	_, err = newTestStore(conf.Storage.Mongo)
	m.Require().NoError(err)
}

func (m *MongoTests) TestNewMongoStorage() {
	r := m.Require()
	conf, err := config.GetConf(testConfigFile)
	if err != nil {
		m.T().Fatalf("Unable to parse config file: %s", err.Error())
	}
	getterSaver, err := NewStorage(conf.Storage.Mongo)

	r.NoError(err)
	r.NotNil(getterSaver.c)
	r.NotNil(getterSaver.d)
	r.NotNil(getterSaver.s)
	r.Equal(getterSaver.url, conf.Storage.Mongo.URL)
}

func TestMongoStorage(t *testing.T) {
	suite.Run(t, new(MongoTests))
}
