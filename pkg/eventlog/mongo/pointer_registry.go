package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/gomods/athens/pkg/eventlog"
)

// Registry is a pointer registry for olypus server event logs
type Registry struct {
	s   *mgo.Session
	d   string // database
	c   string // collection
	url string
}

// NewRegistry creates a pointer registry from backing mongo database
func NewRegistry(url string) (*Registry, error) {
	return NewRegistryWithCollection(url, "pointer-registry")
}

// NewRegistryWithCollection creates a registry using the collection provided
func NewRegistryWithCollection(url, collection string) (*Registry, error) {
	r := Registry{
		url: url,
		c:   collection,
		d:   "athens",
	}
	return &r, r.Connect()
}

// Connect establishes a session with the mongo cluster
func (r *Registry) Connect() error {
	s, err := mgo.Dial(r.url)
	if err != nil {
		return err
	}
	r.s = s

	return nil
}

// LookupPointer returns the pointer to the given deploymentID eventlog
func (r *Registry) LookupPointer(deploymentID string) (string, error) {
	var pointer string

	c := r.s.DB(r.d).C(r.c)
	err := c.FindId(deploymentID).One(&pointer)
	if err == mgo.ErrNotFound {
		return pointer, eventlog.ErrDeploymentNotFound
	}

	return pointer, nil
}
