package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/gomods/athens/pkg/storage/mongoutil"
)

// MetadataStore represents a Mongo backed metadata store.
type MetadataStore struct {
	session     *mgo.Session
	connDetails *mongoutil.ConnDetails
	db          string
	col         string
}

// NewStorage returns an unconnected Mongo backed storage
// that satisfies the Storage interface.  You must call
// Connect() on the returned store before using it.
func NewStorage(connDetails *mongoutil.ConnDetails, dbName string) *MetadataStore {
	return &MetadataStore{connDetails: connDetails, db: dbName}
}

// Connect conntect the the newly created mongo backend.
func (m *MetadataStore) Connect() error {
	conn, err := mongoutil.GetSession(m.connDetails, m.db)
	if err != nil {
		return err
	}
	m.session = conn

	m.col = "cdn_metadata"

	index := mgo.Index{
		Key:        []string{"base_url", "module"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c := m.session.DB(m.db).C(m.col)
	return c.EnsureIndex(index)
}
