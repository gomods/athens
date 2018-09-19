package mongo

import (
	"path/filepath"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/stretchr/testify/suite"
)

var (
	testConfigFile = filepath.Join("..", "..", "..", "config.test.toml")
)

type MongoTests struct {
	suite.Suite
}

func (m *MongoTests) SetupTest() {
	conf := config.GetConfLogErr(testConfigFile, m.T())

	ms, err := newTestStore(conf.Storage.Mongo)
	m.Require().NoError(err)

	ms.s.DB(ms.d).C(ms.c).RemoveAll(nil)
}

func (m *MongoTests) TestNewMongoStorage() {
	r := m.Require()
	conf := config.GetConfLogErr(testConfigFile, m.T())
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
