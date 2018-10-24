package pop

import (
	"fmt"

	"github.com/gobuffalo/pop/logging"
	"github.com/pkg/errors"
)

// CreateDB creates a database, given a connection definition
func CreateDB(c *Connection) error {
	deets := c.Dialect.Details()
	if deets.Database != "" {
		log(logging.Info, fmt.Sprintf("create %s (%s)", deets.Database, c.URL()))
		return errors.Wrapf(c.Dialect.CreateDB(), "couldn't create database %s", deets.Database)
	}
	return nil
}

// DropDB drops an existing database, given a connection definition
func DropDB(c *Connection) error {
	deets := c.Dialect.Details()
	if deets.Database != "" {
		log(logging.Info, fmt.Sprintf("drop %s (%s)", deets.Database, c.URL()))
		return errors.Wrapf(c.Dialect.DropDB(), "couldn't drop database %s", deets.Database)
	}
	return nil
}
