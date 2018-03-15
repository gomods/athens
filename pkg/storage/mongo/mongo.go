package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/gomods/athens/pkg/storage"
)

type storageImpl struct {
	s   *mgo.Session
	d   string // database
	c   string // collection
	url string
}

// NewMongoStorage  returns an unconnected Mongo Module Storage
// that satisfies the Storage interface.  You must call
// Connect() on the returned store before using it.
func NewMongoStorage(url string) storage.StorageConnector {
	return &storageImpl{url: url}
}

func (m *storageImpl) Connect() error {
	s, err := mgo.Dial(m.url)
	if err != nil {
		return err
	}
	m.s = s

	// TODO: database and collection as env vars, or params to New()? together with user/mongo
	m.d = "athens"
	m.c = "modules"

	index := mgo.Index{
		Key:        []string{"BaseURL", "Name", "Version"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c := m.s.DB(m.d).C(m.c)
	return c.EnsureIndex(index)
}
