package mongo

import (
	"errors"
	"strings"

	"github.com/gomods/athens/pkg/storage/mongoutil"
	"github.com/gomods/athens/pkg/user"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// UserStore represents a UserStore implementation backed by mongo.
type UserStore struct {
	s           *mgo.Session
	d           string // database
	c           string // collection
	connDetails *mongoutil.ConnDetails
}

// NewUserStore returns an unconnected UserStore
// that satisfies the UserStore interface.  You must call
// Connect() on the returned store before using it.
//
// TODO: take in database and coll names as params
func NewUserStore(connDetails *mongoutil.ConnDetails) *UserStore {
	return &UserStore{
		connDetails: connDetails,
		d:           "athens",
		c:           "users",
	}
}

// Connect establishes a session to the mongo cluster.
func (m *UserStore) Connect() error {
	s, err := mongoutil.GetSession(m.connDetails, m.d)
	if err != nil {
		panic(err)
	}
	m.s = s

	index := mgo.Index{
		Key:        []string{"provider", "userid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c := m.s.DB(m.d).C(m.c)
	err = c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
	return err
}

// Get returns a user from the Mongo Store
func (m *UserStore) Get(id, provider string) (*user.User, error) {
	c := m.s.DB(m.d).C(m.c)
	result := &user.User{}
	err := c.Find(bson.M{"provider": provider, "userid": id}).One(result)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			err = user.ErrNotFound
		}
	}
	return result, err
}

// Save adds a user to the Mongo Store
func (m *UserStore) Save(u *user.User) error {
	c := m.s.DB(m.d).C(m.c)
	return c.Insert(u)
}

// Update updates a user in the Mongo Store
func (m *UserStore) Update(*user.User) error {
	return errors.New("not implemented")
}
