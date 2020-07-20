package module

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"

	"github.com/gobuffalo/envy"
	"github.com/gomods/athens/pkg/auth"
	"github.com/gomods/athens/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

func (s *ModuleSuite) TestNewGoGetFetcher() {
	r := s.Require()
	fetcher, err := NewGoGetFetcher(s.goBinaryName, "", s.env, s.fs, false, nil)
	r.NoError(err)
	_, ok := fetcher.(*goGetFetcher)
	r.True(ok)
}

func (s *ModuleSuite) TestGoGetFetcherError() {
	fetcher, err := NewGoGetFetcher("invalidpath", "", s.env, afero.NewOsFs(), false, nil)

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
	fetcher, err := NewGoGetFetcher(s.goBinaryName, "", s.env, afero.NewOsFs(), false, nil)
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

func TestGoGetFetcherFetchPrivate(t *testing.T) {
	token := os.Getenv("PROPAGATE_AUTH_TEST_TOKEN")
	if token == "" {
		t.SkipNow()
	}
	var tests = []struct {
		name     string
		desc     string
		prop     bool
		patterns []string
		auth     auth.BasicAuth
		hasErr   bool
		preTest  func(t *testing.T, fetcher Fetcher)
	}{
		{
			name:   "private no token",
			desc:   "cannot fetch a private repository without a basic auth token",
			auth:   auth.BasicAuth{User: "", Password: ""},
			prop:   true,
			hasErr: true,
		},
		{
			name:     "prive fetch",
			desc:     "can successfully download private repository with a valid auth header",
			prop:     true,
			patterns: []string{"*"},
			auth: auth.BasicAuth{
				User:     "athensuser",
				Password: token,
			},
		},
		{
			name: "disable propagation",
			desc: "cannot fetch a private repository even if basic auth is provided when propgate option is false",
			auth: auth.BasicAuth{
				User:     "athensuser",
				Password: token,
			},
			prop:     false,
			patterns: []string{"*"},
			hasErr:   true,
		},
		{
			name: "mismatched auth patterns",
			desc: "cannot fetch a private repository unless the module matches the provided patterns",
			auth: auth.BasicAuth{
				User:     "athensuser",
				Password: token,
			},
			prop:     true,
			patterns: []string{"github.com/other-artifactes/*"},
			hasErr:   true,
		},
		{
			name:     "consecutive private fetch",
			desc:     "this test ensures that the .netrc is removed after a private fetch so credentials are not leakaed to proceeding requests",
			prop:     true,
			patterns: []string{"*"},
			auth:     auth.BasicAuth{},
			preTest: func(t *testing.T, fetcher Fetcher) {
				a := auth.BasicAuth{
					User:     "athensuser",
					Password: token,
				}
				ctx := auth.SetAuthInContext(ctx, a)
				ver, err := fetcher.Fetch(ctx, privateRepoURI, privateRepoVersion)
				require.NoError(t, err)
				require.NoError(t, ver.Zip.Close())
			},
			hasErr: true,
		},
	}
	goBinaryPath := envy.Get("GO_BINARY_PATH", "go")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fetcher, err := NewGoGetFetcher(
				goBinaryPath,
				"",
				[]string{"GOPROXY=direct", "GONOSUMDB=github.com/athens-artifacts/private"},
				afero.NewOsFs(),
				tc.prop,
				tc.patterns,
			)
			require.NoError(t, err)
			if tc.preTest != nil {
				tc.preTest(t, fetcher)
			}
			ctx := auth.SetAuthInContext(ctx, tc.auth)
			ver, err := fetcher.Fetch(ctx, privateRepoURI, privateRepoVersion)
			if tc.hasErr {
				if err == nil {
					ver.Zip.Close()
					t.Fatal("expected an error but got nil")
				}
				return
			}
			require.NoError(t, err)
			defer ver.Zip.Close()

			require.True(t, len(ver.Info) > 0)

			require.True(t, len(ver.Mod) > 0)

			zipBytes, err := ioutil.ReadAll(ver.Zip)
			require.NoError(t, err)
			require.True(t, len(zipBytes) > 0)

			// close the version's zip file (which also cleans up the underlying GOPATH) and expect it to fail again
			require.NoError(t, ver.Zip.Close())
		})
	}
}

func (s *ModuleSuite) TestNotFoundFetches() {
	r := s.Require()
	fetcher, err := NewGoGetFetcher(s.goBinaryName, "", s.env, afero.NewOsFs(), false, nil)
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

	fetcher, err := NewGoGetFetcher(s.goBinaryName, "", []string{"GOPROXY=" + proxyAddr}, afero.NewOsFs(), false, nil)
	r.NoError(err)
	_, err = fetcher.Fetch(ctx, "mockmod.xyz", "v1.2.3")
	if err == nil {
		s.T().Fatal("expected a gosum error but got nil")
	}
	fetcher, err = NewGoGetFetcher(s.goBinaryName, "", []string{"GONOSUMDB=mockmod.xyz", "GOPROXY=" + proxyAddr}, afero.NewOsFs(), false, nil)
	r.NoError(err)
	_, err = fetcher.Fetch(ctx, "mockmod.xyz", "v1.2.3")
	r.NoError(err, "expected the go sum to not be consulted but got an error")
}

func (s *ModuleSuite) TestGoGetDir() {
	r := s.Require()
	t := s.T()
	dir, err := ioutil.TempDir("", "nested")
	r.NoError(err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	fetcher, err := NewGoGetFetcher(s.goBinaryName, dir, s.env, afero.NewOsFs(), false, nil)
	r.NoError(err)

	ver, err := fetcher.Fetch(ctx, repoURI, version)
	r.NoError(err)
	defer ver.Zip.Close()

	dirInfo, err := ioutil.ReadDir(dir)
	r.NoError(err)

	if len(dirInfo) <= 0 {
		t.Fatalf("expected the directory %q to have eat least one sub directory but it was empty", dir)
	}
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
