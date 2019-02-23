package module

import (
	"context"
	"io/ioutil"
	"runtime"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()

func (s *ModuleSuite) TestNewGoGetFetcher() {
	r := s.Require()
	fetcher, err := NewGoGetFetcher(s.goBinaryName, s.fs)
	r.NoError(err)
	_, ok := fetcher.(*goGetFetcher)
	r.True(ok)
}

func (s *ModuleSuite) TestGoGetFetcherError() {
	fetcher, err := NewGoGetFetcher("invalidpath", afero.NewOsFs())

	assert.Nil(s.T(), fetcher)
	if runtime.GOOS == "windows" {
		assert.EqualError(s.T(), err, "exec: \"invalidpath\": executable file not found in %PATH%")
	} else {
		assert.EqualError(s.T(), err, "exec: \"invalidpath\": executable file not found in $PATH")
	}
}

func (s *ModuleSuite) TestGoGetFetcherFetch() {
	r := s.Require()
	// we need to use an OS filesystem because fetch executes vgo on the command line, which
	// always writes to the filesystem
	fetcher, err := NewGoGetFetcher(s.goBinaryName, afero.NewOsFs())
	r.NoError(err)
	ver, err := fetcher.Fetch(ctx, repoURI, version)
	r.NoError(err)
	defer ver.Zip.Reader.Close()

	r.True(len(ver.Info) > 0)

	r.True(len(ver.Mod) > 0)

	zipBytes, err := ioutil.ReadAll(ver.Zip.Reader)
	r.NoError(err)
	r.True(len(zipBytes) > 0)

	// close the version's zip file (which also cleans up the underlying GOPATH) and expect it to fail again
	r.NoError(ver.Zip.Reader.Close())
}
