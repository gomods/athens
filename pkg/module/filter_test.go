package module

import (
	"path/filepath"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/stretchr/testify/suite"
)

const (
	testConfigFile = "../../config.test.toml"
)

type FilterTests struct {
	suite.Suite
	filterFile string
}

func getConf(t *testing.T) *config.Config {
	absPath, err := filepath.Abs(testConfigFile)
	if err != nil {
		t.Errorf("Unable to construct absolute path to test config file")
	}
	conf, err := config.ParseConfigFile(absPath)
	if err != nil {
		t.Errorf("Unable to parse config file")
	}
	return conf
}

func Test_Filter(t *testing.T) {
	conf := getConf(t)
	absPath, err := filepath.Abs(conf.FilterFile)
	if err != nil {
		t.Errorf("Unable to construct absolute path to test config file")
	}
	suite.Run(t, &FilterTests{
		filterFile: absPath,
	})
}

func (t *FilterTests) Test_IgnoreSimple() {
	r := t.Require()

	f, err := NewFilter(t.filterFile)
	if err != nil {
		t.Error(err)
	}
	f.AddRule("github.com/a/b", Exclude)

	r.Equal(Include, f.Rule("github.com/a"))
	r.Equal(Exclude, f.Rule("github.com/a/b"))
	r.Equal(Exclude, f.Rule("github.com/a/b/c"))
	r.Equal(Include, f.Rule("github.com/d"))
	r.Equal(Include, f.Rule("bitbucket.com/a/b"))
}

func (t *FilterTests) Test_IgnoreParentAllowChildren() {
	r := t.Require()

	f, err := NewFilter(t.filterFile)
	if err != nil {
		t.Error(err)
	}
	f.AddRule("github.com/a/b", Exclude)
	f.AddRule("github.com/a/b/c", Include)

	r.Equal(Include, f.Rule("github.com/a"))
	r.Equal(Exclude, f.Rule("github.com/a/b"))
	r.Equal(Include, f.Rule("github.com/a/b/c"))
	r.Equal(Include, f.Rule("github.com/d"))
	r.Equal(Include, f.Rule("bitbucket.com/a/b"))
}

func (t *FilterTests) Test_OnlyAllowed() {
	r := t.Require()

	f, err := NewFilter(t.filterFile)
	if err != nil {
		t.Error(err)
	}
	f.AddRule("github.com/a/b", Include)
	f.AddRule("", Exclude)

	r.Equal(Exclude, f.Rule("github.com/a"))
	r.Equal(Include, f.Rule("github.com/a/b"))
	r.Equal(Include, f.Rule("github.com/a/b/c"))
	r.Equal(Exclude, f.Rule("github.com/d"))
	r.Equal(Exclude, f.Rule("bitbucket.com/a/b"))
}

func (t *FilterTests) Test_Direct() {
	r := t.Require()

	f, err := NewFilter(t.filterFile)
	t.NotNil(f)
	t.NoError(err)
	f.AddRule("github.com/a/b/c", Exclude)
	f.AddRule("github.com/a/b", Direct)
	f.AddRule("github.com/a", Include)
	f.AddRule("", Exclude)

	r.Equal(Include, f.Rule("github.com/a"))
	r.Equal(Direct, f.Rule("github.com/a/b"))
	r.Equal(Exclude, f.Rule("github.com/a/b/c/d"))
}
