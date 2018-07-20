package module

import (
	"io/ioutil"
)

func (s *ModuleSuite) TestNewGoGetFetcher() {
	r := s.Require()
	fs := s.fs
	fetcher, err := NewGoGetFetcher(fs, repoURI, version)
	r.NoError(err)
	goGetFetcher, ok := fetcher.(*goGetFetcher)
	r.True(ok)
	r.Equal(repoURI, goGetFetcher.repoURI)
	r.Equal(version, goGetFetcher.version)
}

func (s *ModuleSuite) TestGoGetFetcherFetch() {
	r := s.Require()
	fs := s.fs
	fetcher, err := NewGoGetFetcher(fs, repoURI, version)
	r.NoError(err)
	ref, err := fetcher.Fetch(repoURI, version)
	r.NoError(err)
	ver, err := ref.Read()
	r.NoError(err)

	r.True(len(ver.Info) > 0)

	r.True(len(ver.Mod) > 0)

	zipBytes, err := ioutil.ReadAll(ver.Zip)
	r.NoError(err)
	r.True(len(zipBytes) > 0)

	r.NoError(ref.Clear())
	ver, err = ref.Read()
	r.NotNil(err)
	r.Nil(ver)
}
