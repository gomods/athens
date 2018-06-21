package rdbms

import (
	"github.com/gobuffalo/pop"
)

// ModuleStore represents a rdbms(postgres, mysql, sqlite, cockroachdb) backed storage backend.
type ModuleStore struct {
	conn *pop.Connection
	// connectionName string // settings name from database.yml
}

// NewRDBMSStorage  returns an unconnected RDBMS Module Storage
// that satisfies the Storage interface. You must call
// Connect() on the returned store before using it.
// connectionName
func NewRDBMSStorage(connectionName string) (*ModuleStore, error) {
	c, err := pop.Connect(connectionName)
	if err != nil {
		return nil, err
	}
	return &ModuleStore{
		// connectionName: connectionName,
		conn: c,
	}, nil

}
