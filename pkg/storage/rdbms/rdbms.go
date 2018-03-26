package rdbms

import (
	"github.com/gobuffalo/pop"
)

// ModuleStore represents a rdbms(postgres, mysql, sqlite, cockroachdb) backed storage backend.
type ModuleStore struct {
	conn *pop.Connection
	e    string
}

// NewRDBMSStorage  returns an unconnected RDBMS Module Storage
// that satisfies the Storage interface. You must call
// Connect() on the returned store before using it.
// connectionName
func NewRDBMSStorage(connectionName string) *ModuleStore {
	return &ModuleStore{
		e: connectionName,
	}
}

// Connect creates connection to rdmbs backend.
func (r *ModuleStore) Connect() error {
	c, err := pop.Connect(r.e)
	if err != nil {
		return err
	}
	r.conn = c
	return nil
}
