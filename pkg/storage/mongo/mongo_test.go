package mongo

import (
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/config/env"
)

func (m *MongoTests) TestNewMongoStorage() {
	// TODO: what is the difference between all_test, mongo_test, test_suite.go??
	r := m.Require()
	muri := env.MongoConnectionString()
	certPath := env.MongoCertPath()
	conf := &config.MongoConfig{
		URL:      muri,
		CertPath: certPath,
		TimeoutConf: config.TimeoutConf{
			Timeout: 1,
		},
	}
	getterSaver, err := NewStorage(conf)

	r.NoError(err)
	r.NotNil(getterSaver.c)
	r.NotNil(getterSaver.d)
	r.NotNil(getterSaver.s)
	r.Equal(getterSaver.url, muri)
}
