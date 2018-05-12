package mongo

import (
	"strings"

	"github.com/gomods/athens/pkg/proxy"
	"github.com/gomods/athens/pkg/user"

	"github.com/globalsign/mgo"
)

// StateStore represents a state.Store implementation backed by mongo.
type StateStore struct {
	s   *mgo.Session
	d   string // database
	c   string // collection
	url string
}

// NewStateStore returns an unconnected StateStore
// that satisfies the state.Store interface.  You must call
// Connect() on the returned store before using it.
func NewStateStore(url string) *StateStore {
	return &StateStore{url: url}
}

// Connect establishes a session to the mongo cluster.
func (m *StateStore) Connect() error {
	s, err := mgo.Dial(m.url)
	if err != nil {
		return err
	}
	m.s = s

	// TODO(BJK) database and collection as env vars, or params to New()?
	m.d = "athens"
	m.c = "proxystate"

	return nil
}

// Get returns a user from the Mongo Store
func (m *StateStore) Get() (proxy.State, error) {
	c := m.s.DB(m.d).C(m.c)
	result := &proxy.State{}
	err := c.Find(nil).One(result)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			err = user.ErrNotFound
		}
	}
	return *result, err
}

// Set updates state in the Mongo Store
func (m *StateStore) Set(endpoint, sequenceID string) error {
	s := &proxy.State{OlympusEndpoint: endpoint, SequenceID: sequenceID}
	c := m.s.DB(m.d).C(m.c)
	_, err := c.Upsert(nil, s)
	return err
}

// Clear clears state in case of olympus goes down
func (m *StateStore) Clear() error {
	c := m.s.DB(m.d).C(m.c)
	_, err := c.RemoveAll(nil)
	return err
}
