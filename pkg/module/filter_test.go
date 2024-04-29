package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

func testConfigFile(t *testing.T) (testConfigFile string) {
	testConfigFile = filepath.Join("..", "..", "config.dev.toml")
	if err := os.Chmod(testConfigFile, 0o700); err != nil {
		t.Fatalf("%s\n", err)
	}
	return testConfigFile
}

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
	f.AddRule("github.com/a/b", nil, Exclude)

	r.Equal(Include, f.Rule("github.com/a", ""))
	r.Equal(Exclude, f.Rule("github.com/a/b", ""))
	r.Equal(Exclude, f.Rule("github.com/a/b/c", ""))
	r.Equal(Include, f.Rule("github.com/d", ""))
	r.Equal(Include, f.Rule("bitbucket.com/a/b", ""))
}

func (t *FilterTests) Test_IgnoreParentAllowChildren() {
	r := t.Require()

	filter := tempFilterFile(t.T())
	defer os.Remove(filter)

	f, err := NewFilter(filter)
	r.NoError(err)
	f.AddRule("github.com/a/b", nil, Exclude)
	f.AddRule("github.com/a/b/c", nil, Include)

	r.Equal(Include, f.Rule("github.com/a", ""))
	r.Equal(Exclude, f.Rule("github.com/a/b", ""))
	r.Equal(Include, f.Rule("github.com/a/b/c", ""))
	r.Equal(Include, f.Rule("github.com/d", ""))
	r.Equal(Include, f.Rule("bitbucket.com/a/b", ""))
}

func (t *FilterTests) Test_OnlyAllowed() {
	r := t.Require()

	filter := tempFilterFile(t.T())
	defer os.Remove(filter)

	f, err := NewFilter(filter)
	r.NoError(err)
	f.AddRule("github.com/a/b", nil, Include)
	f.AddRule("", nil, Exclude)

	r.Equal(Exclude, f.Rule("github.com/a", ""))
	r.Equal(Include, f.Rule("github.com/a/b", ""))
	r.Equal(Include, f.Rule("github.com/a/b/c", ""))
	r.Equal(Exclude, f.Rule("github.com/d", ""))
	r.Equal(Exclude, f.Rule("bitbucket.com/a/b", ""))
}

func (t *FilterTests) Test_Direct() {
	r := t.Require()

	filter := tempFilterFile(t.T())
	defer os.Remove(filter)

	f, err := NewFilter(filter)
	r.NoError(err)
	f.AddRule("github.com/a/b/c", nil, Exclude)
	f.AddRule("github.com/a/b", nil, Direct)
	f.AddRule("github.com/a", nil, Include)
	f.AddRule("", nil, Exclude)

	r.Equal(Include, f.Rule("github.com/a", ""))
	r.Equal(Direct, f.Rule("github.com/a/b", ""))
	r.Equal(Exclude, f.Rule("github.com/a/b/c/d", ""))
}

func (t *FilterTests) Test_versionFilter() {
	r := t.Require()
	filter := tempFilterFile(t.T())
	defer os.Remove(filter)

	f, err := NewFilter(filter)
	r.NoError(err)
	f.AddRule("", nil, Exclude)
	f.AddRule("github.com/a/b", []string{"v1."}, Include)
	f.AddRule("github.com/a/b/c", []string{"v1.2.", "v0.8."}, Direct)

	r.Equal(Exclude, f.Rule("github.com/d/e", "v1.2.0"))
	r.Equal(Exclude, f.Rule("github.com/a/b", "v10.0.0"))
	r.Equal(Include, f.Rule("github.com/a/b", "v1.5.0"))
	r.Equal(Direct, f.Rule("github.com/a/b/c/d", "v1.2.3"))
	r.Equal(Include, f.Rule("github.com/a/b/c/d", "v1.3.4"))
}

func (t *FilterTests) Test_versionFilterMinor() {
	r := t.Require()
	filter := tempFilterFile(t.T())
	defer os.Remove(filter)

	f, err := NewFilter(filter)
	r.NoError(err)
	f.AddRule("", nil, Exclude)
	f.AddRule("github.com/a/b", []string{"~v1.2.3", "~v2.3.40"}, Include)
	r.Equal(Include, f.Rule("github.com/a/b", "v1.2.3"))
	r.Equal(Include, f.Rule("github.com/a/b", "v1.2.5"))
	r.Equal(Exclude, f.Rule("github.com/a/b", "v1.2.2"))
	r.Equal(Exclude, f.Rule("github.com/a/b", "v1.3.3"))
	r.Equal(Include, f.Rule("github.com/a/b", "v2.3.45"))
	r.Equal(Exclude, f.Rule("github.com/a/b", "v2.2.45"))
	r.Equal(Exclude, f.Rule("github.com/a/b", "v2.3.20"))
}

func (t *FilterTests) Test_versionFilterMiddle() {
	r := t.Require()
	filter := tempFilterFile(t.T())
	defer os.Remove(filter)

	f, err := NewFilter(filter)
	r.NoError(err)
	f.AddRule("", nil, Exclude)
	f.AddRule("github.com/a/b", []string{"^v1.2.3", "^v2.3.40"}, Include)
	r.Equal(Include, f.Rule("github.com/a/b", "v1.2.3"))
	r.Equal(Include, f.Rule("github.com/a/b", "v1.2.5"))
	r.Equal(Include, f.Rule("github.com/a/b", "v1.4.2"))
	r.Equal(Exclude, f.Rule("github.com/a/b", "v1.2.1"))
	r.Equal(Include, f.Rule("github.com/a/b", "v2.3.45"))
	r.Equal(Include, f.Rule("github.com/a/b", "v2.4.1"))
	r.Equal(Exclude, f.Rule("github.com/a/b", "v2.2.45"))
}

func (t *FilterTests) Test_versionFilterLess() {
	r := t.Require()
	filter := tempFilterFile(t.T())
	defer os.Remove(filter)

	f, err := NewFilter(filter)
	r.NoError(err)
	f.AddRule("", nil, Exclude)
	f.AddRule("github.com/a/b", []string{"<v2.3.40"}, Include)
	r.Equal(Include, f.Rule("github.com/a/b", "v1.2.3"))
	r.Equal(Include, f.Rule("github.com/a/b", "v1.4.2"))
	r.Equal(Include, f.Rule("github.com/a/b", "v1.2.1"))
	r.Equal(Include, f.Rule("github.com/a/b", "v2.3.39"))
	r.Equal(Include, f.Rule("github.com/a/b", "v2.2.45"))
	r.Equal(Exclude, f.Rule("github.com/a/b", "v2.4.1"))
}

func (t *FilterTests) Test_versionFilterRobust() {
	r := t.Require()
	filter := tempFilterFile(t.T())
	defer os.Remove(filter)

	f, err := NewFilter(filter)
	r.NoError(err)
	f.AddRule("", nil, Exclude)
	f.AddRule("github.com/a/b", []string{"abcd"}, Include)
	f.AddRule("github.com/c/d", []string{"e"}, Include)

	r.Equal(Exclude, f.Rule("github.com/a/b", "a"))
	r.Equal(Exclude, f.Rule("github.com/c/d", "fg"))
}

func (t *FilterTests) Test_initFromConfig() {
	r := t.Require()
	filterFile := tempFilterFile(t.T())
	defer os.Remove(filterFile)

	goodInput := []byte("+ github.com/a/b\n\n# some comment\n- github.com/c/d\n\nD github.com/x")
	os.WriteFile(filterFile, goodInput, 0o644)

	f, err := initFromConfig(filterFile)
	r.NotNil(f)
	r.NoError(err)

	badInput := []byte("+ github.com/a/b\n\n# some comment\n\n- github.com/c/d\n\nD github.com/x\nsome_random_line")
	os.WriteFile(filterFile, badInput, 0o644)
	f, err = initFromConfig(filterFile)
	r.Nil(f)
	r.Error(err)

	versionInput := []byte("+ github.com/a/b\n\n# some comment\n\n- github.com/c/d v1,v2.3.4,v3.2.*\n\nD github.com/x\n")
	os.WriteFile(filterFile, versionInput, 0o644)
	f, err = initFromConfig(filterFile)
	r.NotNil(f)
	r.NoError(err)
}

func tempFilterFile(t *testing.T) (path string) {
	filter, err := os.CreateTemp(os.TempDir(), "filter-")
	if err != nil {
		t.FailNow()
	}
	return filter.Name()
}
