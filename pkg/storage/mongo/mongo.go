package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/gomods/athens/pkg/storage/mongoutil"
)

const (
	athensDBName    = "athens"
	modulesCollName = "modules"
)

// ModuleStore represents a mongo backed storage backend.
type ModuleStore struct {
	s           *mgo.Session
	d           string // database
	c           string // collection
	connDetails *mongoutil.ConnDetails
}

// NewStorage returns an unconnected Mongo backed storage
// that satisfies the Backend interface.  You must call
// Connect() on the returned store before using it.
//
// TODO: take the database and collection names as parameters
func NewStorage(connDetails *mongoutil.ConnDetails) *ModuleStore {
	return &ModuleStore{connDetails: connDetails}
}

// Connect connects the the newly created mongo backend.
func (m *ModuleStore) Connect() error {
	s, err := mongoutil.GetSession(m.connDetails, athensDBName)
	if err != nil {
		return err
	}
	m.s = s

	index := mgo.Index{
		Key:        []string{"base_url", "module", "version"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c := m.s.DB(athensDBName).C(modulesCollName)
	return c.EnsureIndex(index)
}
