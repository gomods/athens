package mongo

import (
	"github.com/gomods/athens/pkg/config"
)

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
