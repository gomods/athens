package module

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gobuffalo/envy"
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
	assert.EqualError(s.T(), err, "exec: \"invalidpath\": executable file not found in $PATH")
}

func (s *ModuleSuite) TestGoGetFetcherFetch() {
	r := s.Require()
	// we need to use an OS filesystem because fetch executes vgo on the command line, which
	// always writes to the filesystem
	fetcher, err := NewGoGetFetcher(s.goBinaryName, afero.NewOsFs())
	r.NoError(err)
	ver, err := fetcher.Fetch(ctx, repoURI, version)
	r.NoError(err)
	defer ver.Zip.Close()

	r.True(len(ver.Info) > 0)

	r.True(len(ver.Mod) > 0)

	zipBytes, err := ioutil.ReadAll(ver.Zip)
	r.NoError(err)
	r.True(len(zipBytes) > 0)

	// close the version's zip file (which also cleans up the underlying diskref's GOPATH) and expect it to fail again
	r.NoError(ver.Zip.Close())
}

func ExampleFetcher() {
	repoURI := "github.com/arschles/assert"
	version := "v1.0.0"
	goBinaryName := envy.Get("GO_BINARY_PATH", "go")
	fetcher, err := NewGoGetFetcher(goBinaryName, afero.NewOsFs())
	if err != nil {
		log.Fatal(err)
	}
	versionData, err := fetcher.Fetch(ctx, repoURI, version)
	// handle errors if any
	if err != nil {
		return
	}
	// Close the handle to versionData.Zip once done
	// This will also handle cleanup so it's important to call Close
	defer versionData.Zip.Close()
	if err != nil {
		return
	}
	// Do something with versionData
	fmt.Println(string(versionData.Mod))
	// Output: module github.com/arschles/assert
}
