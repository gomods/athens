package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/gomods/athens/pkg/storage/mongo/conn"
)

const (
	// TODO: database and collection as env vars, or params to New()? together with user/mongo
	athensDB          = "athens"
	modulesCollection = "modules"
)

// ModuleStore represents a mongo backed storage backend.
type ModuleStore struct {
	sess        *mgo.Session
	connDetails *conn.Details
}

// NewStorage returns an unconnected Mongo backed storage
// that satisfies the Backend interface.  You must call
// Connect() on the returned store before using it.
func NewStorage(connDetails *conn.Details) *ModuleStore {
	return &ModuleStore{connDetails: connDetails}
}

// Connect conntect the the newly created mongo backend.
func (m *ModuleStore) Connect() error {
	s, err := conn.NewSession(m.connDetails, athensDB)
	if err != nil {
		return err
	}
	m.sess = s

	index := mgo.Index{
		Key:        []string{"base_url", "module", "version"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c := m.sess.DB(athensDB).C(modulesCollection)
	return c.EnsureIndex(index)
}
