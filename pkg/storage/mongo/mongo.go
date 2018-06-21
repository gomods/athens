package mongo

import (
	"github.com/globalsign/mgo"
)

const (
	coll = "modules"
	db   = "athens"
)

// ModuleStore represents a mongo backed storage backend.
type ModuleStore struct {
	s     *mgo.Session
	deets *ConnDetails
	d     string // database
	c     string // collection
	url   string
}

// NewStorage returns an unconnected Mongo backed storage
// that satisfies the Backend interface.  You must call
// Connect() on the returned store before using it.
func NewStorage(deets *ConnDetails) (*ModuleStore, error) {
	sess, err := GetSession(deets, db)
	if err != nil {
		return nil, err
	}
	ms := &ModuleStore{s: sess, deets: deets}

	index := mgo.Index{
		Key:        []string{"base_url", "module", "version"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c := sess.DB(db).C(coll)
	if err := c.EnsureIndex(index); err != nil {
		return nil, err
	}
	return ms, nil
}
