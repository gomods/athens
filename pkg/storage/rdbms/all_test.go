package rdbms

import (
	"testing"

	"github.com/gomods/athens/pkg/storage"
	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/actions"
)

const (
	module  = "testmodule"
	version = "v1.0.0"
)

var (
	// TODO: put these values inside of the suite, and generate longer values.
	// This should help catch edge cases, like https://github.com/gomods/athens/issues/38
	//
	// Also, consider doing something similar to what testing/quick does
	// with the Generator interface (https://godoc.org/testing/quick#Generator).
	// The rough, simplified idea would be to run a single test case multiple
	// times over different (increasing) values.
	mod = []byte("123")
	zip = []byte("456")
)

type RDBMSTests struct {
	*suite.Action
	storage storage.StorageConnector
}

func (rd *RDBMSTests) SetupTest() {
	store := NewRDBMSStorage("development_postgres")
	store.Connect()
	store.conn.TruncateAll()
	rd.storage = store
}

func TestDiskStorage(t *testing.T) {
	suite.Run(t, new(RDBMSTests))
}

func Test_ActionSuite(t *testing.T) {
	as := &RDBMSTests{suite.NewAction(actions.App())}
	suite.Run(t, as)
}