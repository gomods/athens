package mongo

import (
	"context"
	"io/ioutil"

	"github.com/bketelsen/buffet"

	"github.com/gobuffalo/buffalo"
)

func (m *MongoTests) TestGetSaveListRoundTrip() {
	c := &buffalo.DefaultContext{
		Context: context.Background(),
	}
	sp := buffet.SpanFromContext(c)
	sp.SetOperationName("test.storage.mongo.GetSaveListRoundTrip")
	defer sp.Finish()

	r := m.Require()
	m.storage.Save(c, module, version, mod, zip, info)
	listedVersions, err := m.storage.List(module)
	r.NoError(err)
	r.Equal(1, len(listedVersions))
	retVersion := listedVersions[0]
	r.Equal(version, retVersion)
	gotten, err := m.storage.Get(module, version)
	r.NoError(err)
	defer gotten.Zip.Close()
	// TODO: test the time
	r.Equal(gotten.Mod, mod)
	zipContent, err := ioutil.ReadAll(gotten.Zip)
	r.NoError(err)
	r.Equal(zipContent, zip)
	r.Equal(gotten.Info, info)
}

func (m *MongoTests) TestNewMongoStorage() {
	r := m.Require()
	url := "mongodb://127.0.0.1:27017"
	getterSaver := NewStorage(url)
	getterSaver.Connect()

	r.NotNil(getterSaver.c)
	r.NotNil(getterSaver.d)
	r.NotNil(getterSaver.s)
	r.Equal(getterSaver.url, url)
}
