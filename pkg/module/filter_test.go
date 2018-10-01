package module

import (
	"path/filepath"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/stretchr/testify/suite"
)

var (
	testConfigFile = filepath.Join("..", "..", "config.dev.toml")
)

type FilterTests struct {
	suite.Suite
}

func Test_Filter(t *testing.T) {
	suite.Run(t, new(FilterTests))
}

func (t *FilterTests) Test_IgnoreSimple() {
	r := t.Require()
	conf := GetConfLogErr(testConfigFile, t.T())
	f := NewFilter(conf.FilterFile)
	f.AddRule("github.com/a/b", Exclude)

	r.Equal(Include, f.Rule("github.com/a"))
	r.Equal(Exclude, f.Rule("github.com/a/b"))
	r.Equal(Exclude, f.Rule("github.com/a/b/c"))
	r.Equal(Include, f.Rule("github.com/d"))
	r.Equal(Include, f.Rule("bitbucket.com/a/b"))
}

func (t *FilterTests) Test_IgnoreParentAllowChildren() {
	r := t.Require()
	conf := GetConfLogErr(testConfigFile, t.T())
	f := NewFilter(conf.FilterFile)
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

	conf := GetConfLogErr(testConfigFile, t.T())
	f := NewFilter(conf.FilterFile)
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

	conf := GetConfLogErr(testConfigFile, t.T())
	f := NewFilter(conf.FilterFile)
	f.AddRule("github.com/a/b/c", Exclude)
	f.AddRule("github.com/a/b", Direct)
	f.AddRule("github.com/a", Include)
	f.AddRule("", Exclude)

	r.Equal(Include, f.Rule("github.com/a"))
	r.Equal(Direct, f.Rule("github.com/a/b"))
	r.Equal(Exclude, f.Rule("github.com/a/b/c/d"))
}

// GetConfLogErr is similar to GetConf, except it logs a failure for the calling test
// if any errors are encountered
func GetConfLogErr(path string, t *testing.T) *config.Config {
	c, err := config.GetConf(path)
	if err != nil {
		t.Fatalf("Unable to parse config file: %s", err.Error())
		return nil
	}
	return c
}
