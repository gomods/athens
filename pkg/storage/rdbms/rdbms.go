package rdbms

import (
	"github.com/gobuffalo/pop"
)

// ModuleStore represents a rdbms(postgres, mysql, sqlite, cockroachdb) backed storage backend.
type ModuleStore struct {
	conn           *pop.Connection
	connectionName string // settings name from database.yml
}

// NewRDBMSStorage  returns an unconnected RDBMS Module Storage
// that satisfies the Storage interface.
func NewRDBMSStorage(connectionName string) (*ModuleStore, error) {
	ms := &ModuleStore{
		connectionName: connectionName,
	}
	err := ms.connect()
	return ms, err
}

// NewRDBMSStorageWithConn  returns a connected RDBMS Module Storage
// that satisfies the Storage interface.
func NewRDBMSStorageWithConn(connection *pop.Connection) (*ModuleStore, error) {
	ms := &ModuleStore{
		conn: connection,
	}
	err := ms.connect()
	return ms, err
}

func (r *ModuleStore) connect() error {
	c, err := pop.Connect(r.connectionName)
	if err != nil {
		return err
	}
	r.conn = c
	return nil
}
