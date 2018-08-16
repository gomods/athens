package actions

import (
	"testing"

	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/mem"
)

type ActionSuite struct {
	*suite.Action
	store storage.Backend
}

func Test_ActionSuite(t *testing.T) {
	store, err := mem.NewStorage()
	if err != nil {
		t.Fatalf("couldn't create a new in-memory storage (%s)", err)
	}
	app, err := App(store)
	if err != nil {
		t.Fatal(err)
	}
	as := &ActionSuite{
		Action: suite.NewAction(app),
		store:  store,
	}
	suite.Run(t, as)
}
