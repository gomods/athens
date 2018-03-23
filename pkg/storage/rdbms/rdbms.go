package rdbms

import (
	"github.com/gobuffalo/pop"
)

type RDBMSModuleStore struct {
	conn *pop.Connection
	e    string
}

// NewRDBMSStorage  returns an unconnected RDBMS Module Storage
// that satisfies the Storage interface. You must call
// Connect() on the returned store before using it.
// connectionName
func NewRDBMSStorage(connectionName string) *RDBMSModuleStore {
	return &RDBMSModuleStore{
		e: connectionName,
	}
}

func (r *RDBMSModuleStore) Connect() error {
	c, err := pop.Connect(r.e)
	if err != nil {
		return err
	}
	r.conn = c
	return nil
}
