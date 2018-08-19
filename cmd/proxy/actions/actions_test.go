package actions

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/config"
)

const (
	testConfigFile = "../../../config.test.toml"
)

func getConf() (*config.Config, error) {
	absPath, err := filepath.Abs(testConfigFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to construct absolute path to test config file")
	}
	conf, err := config.ParseConfigFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse test config file: %s", err.Error())
	}
	return conf, nil
}

type ActionSuite struct {
	*suite.Action
}

func Test_ActionSuite(t *testing.T) {
	conf, err := getConf()
	if err != nil {
		t.Fatal(err)
	}
	app, err := App(conf)
	if err != nil {
		t.Fatal(err)
	}
	as := &ActionSuite{suite.NewAction(app)}
	suite.Run(t, as)
}
