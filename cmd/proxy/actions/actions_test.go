package actions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/config"
)

func testConfigFile(t *testing.T) (testConfigFile string) {
	testConfigFile = filepath.Join("..", "..", "..", "config.dev.toml")
	if err := os.Chmod(testConfigFile, 0700); err != nil {
		t.Fatalf("%s\n", err)
	}
	return testConfigFile
}

type ActionSuite struct {
	*suite.Action
}

func Test_ActionSuite(t *testing.T) {
	conf, err := config.GetConf(testConfigFile(t))
	if err != nil {
		t.Fatalf("Unable to parse config file: %s", err.Error())
	}
	app, err := App(conf)
	if err != nil {
		t.Fatal(err)
	}
	as := &ActionSuite{suite.NewAction(app)}
	suite.Run(t, as)
}
