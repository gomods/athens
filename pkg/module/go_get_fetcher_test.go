package module

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"

	"github.com/gomods/athens/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()

func (s *ModuleSuite) TestNewGoGetFetcher() {
	r := s.Require()
	fetcher, err := NewGoGetFetcher(s.goBinaryName, s.env, s.fs)
	r.NoError(err)
	_, ok := fetcher.(*goGetFetcher)
	r.True(ok)
}

func (s *ModuleSuite) TestGoGetFetcherError() {
	fetcher, err := NewGoGetFetcher("invalidpath", s.env, afero.NewOsFs())

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
	fetcher, err := NewGoGetFetcher(s.goBinaryName, s.env, afero.NewOsFs())
	r.NoError(err)
	ver, err := fetcher.Fetch(ctx, repoURI, version)
	r.NoError(err)
	defer ver.Zip.Close()

	r.True(len(ver.Info) > 0)

	r.True(len(ver.Mod) > 0)

	zipBytes, err := ioutil.ReadAll(ver.Zip)
	r.NoError(err)
	r.True(len(zipBytes) > 0)

	// close the version's zip file (which also cleans up the underlying GOPATH) and expect it to fail again
	r.NoError(ver.Zip.Close())
}

func (s *ModuleSuite) TestNotFoundFetches() {
	r := s.Require()
	fetcher, err := NewGoGetFetcher(s.goBinaryName, s.env, afero.NewOsFs())
	r.NoError(err)
	// when someone buys laks47dfjoijskdvjxuyyd.com, and implements
	// a git server on top of it, this test will fail :)
	_, err = fetcher.Fetch(ctx, "laks47dfjoijskdvjxuyyd.com/pkg/errors", "v0.8.1")
	if err == nil {
		s.Fail("expected an error but got nil")
	}
	if errors.Kind(err) != errors.KindNotFound {
		s.Failf("incorrect error kind", "expected a not found error but got %v", errors.Kind(err))
	}
}

func (s *ModuleSuite) TestGoGetFetcherSumDB() {
	if os.Getenv("SKIP_UNTIL_113") != "" {
		return
	}
	r := s.Require()
	zipBytes, err := ioutil.ReadFile("test_data/mockmod.xyz@v1.2.3.zip")
	r.NoError(err)
	mp := &mockProxy{paths: map[string][]byte{
		"/mockmod.xyz/@v/v1.2.3.info": []byte(`{"Version":"v1.2.3"}`),
		"/mockmod.xyz/@v/v1.2.3.mod":  []byte(`{"module mod}`),
		"/mockmod.xyz/@v/v1.2.3.zip":  zipBytes,
	}}
	proxyAddr, close := s.getProxy(mp)
	defer close()

	fetcher, err := NewGoGetFetcher(s.goBinaryName, []string{"GOPROXY=" + proxyAddr}, afero.NewOsFs())
	r.NoError(err)
	_, err = fetcher.Fetch(ctx, "mockmod.xyz", "v1.2.3")
	if err == nil {
		s.T().Fatal("expected a gosum error but got nil")
	}
	fetcher, err = NewGoGetFetcher(s.goBinaryName, []string{"GONOSUMDB=mockmod.xyz", "GOPROXY=" + proxyAddr}, afero.NewOsFs())
	r.NoError(err)
	_, err = fetcher.Fetch(ctx, "mockmod.xyz", "v1.2.3")
	r.NoError(err, "expected the go sum to not be consulted but got an error")
}

func (s *ModuleSuite) getProxy(h http.Handler) (addr string, close func()) {
	srv := httptest.NewServer(h)
	return srv.URL, srv.Close
}

type mockProxy struct {
	paths map[string][]byte
}

func (m *mockProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp, ok := m.paths[r.URL.Path]
	if !ok {
		w.WriteHeader(404)
		return
	}
	w.Write(resp)
}
