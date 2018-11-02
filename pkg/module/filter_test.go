package module

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

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

func (t *FilterTests) Test_NewFilter() {
	r := t.Require()
	mf, err := NewFilter("")
	r.NoError(err, "When a file name is empty string return no error")
	r.Nil(nil, mf)

	mf, err = NewFilter("nofile")
	r.Nil(mf)
	r.Error(err)

	filter := tempFilterFile(t.T())
	defer os.Remove(filter)

	mf, err = NewFilter(filter)
	r.Equal(filter, mf.filePath)
	r.NoError(err)

}

func (t *FilterTests) Test_IgnoreSimple() {
	r := t.Require()

	filter := tempFilterFile(t.T())
	defer os.Remove(filter)

	f, err := NewFilter(filter)
	r.NoError(err)
	f.AddRule("github.com/a/b", Exclude)

	r.Equal(Include, f.Rule("github.com/a"))
	r.Equal(Exclude, f.Rule("github.com/a/b"))
	r.Equal(Exclude, f.Rule("github.com/a/b/c"))
	r.Equal(Include, f.Rule("github.com/d"))
	r.Equal(Include, f.Rule("bitbucket.com/a/b"))
}

func (t *FilterTests) Test_IgnoreParentAllowChildren() {
	r := t.Require()

	filter := tempFilterFile(t.T())
	defer os.Remove(filter)

	f, err := NewFilter(filter)
	r.NoError(err)
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

	filter := tempFilterFile(t.T())
	defer os.Remove(filter)

	f, err := NewFilter(filter)
	r.NoError(err)
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

	filter := tempFilterFile(t.T())
	defer os.Remove(filter)

	f, err := NewFilter(filter)
	r.NoError(err)
	f.AddRule("github.com/a/b/c", Exclude)
	f.AddRule("github.com/a/b", Direct)
	f.AddRule("github.com/a", Include)
	f.AddRule("", Exclude)

	r.Equal(Include, f.Rule("github.com/a"))
	r.Equal(Direct, f.Rule("github.com/a/b"))
	r.Equal(Exclude, f.Rule("github.com/a/b/c/d"))
}

func (t *FilterTests) Test_initFromConfig() {
	r := t.Require()
	filterFile := tempFilterFile(t.T())
	defer os.Remove(filterFile)

	goodInput := []byte("+ github.com/a/b\n\n# some comment\n- github.com/c/d\n\nD github.com/x")
	ioutil.WriteFile(filterFile, goodInput, 0644)

	f, err := initFromConfig(filterFile)
	r.NotNil(f)
	r.NoError(err)

	badInput := []byte("+ github.com/a/b\n\n# some comment\n\n- github.com/c/d\n\nD github.com/x\nsome_random_line")
	ioutil.WriteFile(filterFile, badInput, 0644)
	f, err = initFromConfig(filterFile)
	r.Nil(f)
	r.Error(err)

}

func tempFilterFile(t *testing.T) (path string) {
	filter, err := ioutil.TempFile(os.TempDir(), "filter-")
	if err != nil {
		t.FailNow()
	}
	return filter.Name()
}
