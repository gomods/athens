package mongo

import (
	"time"

	"github.com/gomods/athens/pkg/config/env"
)

func (m *MongoTests) TestNewMongoStorage() {
	// TODO: what is the difference between all_test, mongo_test, test_suite.go??
	r := m.Require()
	muri := env.MongoConnectionString()
	certPath := env.MongoCertPath()
	getterSaver, err := NewStorageWithCert(muri, certPath, time.Second)

	r.NoError(err)
	r.NotNil(getterSaver.c)
	r.NotNil(getterSaver.d)
	r.NotNil(getterSaver.s)
	r.Equal(getterSaver.url, muri)
}
