package mongo

import (
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"github.com/gomods/athens/pkg/errors"
)

// ModuleStore represents a mongo backed storage backend.
type ModuleStore struct {
	s       *mgo.Session
	d       string // database
	c       string // collection
	url     string
	timeout time.Duration
}

// NewStorage returns an connected Mongo backed storage
// that satisfies the Backend interface.
func NewStorage(url string, timeout time.Duration) (*ModuleStore, error) {
	const op errors.Op = "mongo.NewStorage"
	ms := &ModuleStore{url: url, timeout: timeout}

	err := ms.connect()
	if err != nil {
		return nil, errors.E(op, err)
	}
	return ms, nil
}

// Connect conntect the the newly created mongo backend.
func (m *ModuleStore) connect() error {
	const op errors.Op = "mongo.Connect"
	s, err := mgo.DialWithTimeout(m.url, m.timeout)
	if err != nil {
		return errors.E(op, err)
	}
	m.s = s

	// TODO: database and collection as env vars, or params to New()? together with user/mongo
	m.d = "athens"
	m.c = "modules"

	index := mgo.Index{
		Key:        []string{"base_url", "module", "version"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c := m.s.DB(m.d).C(m.c)
	return c.EnsureIndex(index)
}

func (m *ModuleStore) gridFileName(mod, ver string) string {
	return strings.Replace(mod, "/", "_", -1) + "_" + ver + ".zip"
}
