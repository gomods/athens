package actions

import (
	"path/filepath"
	"testing"

	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/config"
)

var (
	testConfigFile = filepath.Join("..", "..", "..", "config.test.toml")
)

type ActionSuite struct {
	*suite.Action
}

func Test_ActionSuite(t *testing.T) {
	conf := config.GetConfLogErr(testConfigFile, t)
	app, err := App(conf)
	if err != nil {
		t.Fatal(err)
	}
	as := &ActionSuite{suite.NewAction(app)}
	suite.Run(t, as)
}
