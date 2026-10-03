package actions

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"

	"github.com/gomods/athens/pkg/build"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type routeTest struct {
	method string
	path   string
	body   string
	test   func(t *testing.T, req *http.Request, resp *http.Response)
}

func TestProxyRoutes(t *testing.T) {
	r := mux.NewRouter()
	s, err := mem.NewStorage()
	require.NoError(t, err)
	l := log.NoOpLogger()
	c, err := config.Load("")
	require.NoError(t, err)
	c.NoSumPatterns = []string{"*"} // catch all patterns with noSumWrapper to ensure the sumdb handler doesn't make a real http request to the sumdb server.
	c.PathPrefix = "/prefix"
	subRouter := r.PathPrefix(c.PathPrefix).Subrouter()
	err = addProxyRoutes(t.Context(), subRouter, s, l, c)
	require.NoError(t, err)

	baseURL := "https://athens.azurefd.net" + c.PathPrefix

	testCases := []routeTest{
		{"GET", "/", "", func(t *testing.T, req *http.Request, resp *http.Response) {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			tmp, err := template.New("home").Parse(homepage)
			assert.NoError(t, err)

			templateData := make(map[string]string)

			templateData["Host"] = req.Host

			if !strings.HasPrefix(templateData["Host"], "http://") && !strings.HasPrefix(templateData["Host"], "https://") {
				if req.TLS != nil {
					templateData["Host"] = "https://" + templateData["Host"]
				} else {
					templateData["Host"] = "http://" + templateData["Host"]
				}
			}

			templateData["NoSumPatterns"] = strings.Join(c.NoSumPatterns, ",")

			var expected strings.Builder
			err = tmp.ExecuteTemplate(&expected, "home", templateData)
			require.NoError(t, err)

			assert.Equal(t, expected.String(), string(body))
		}},
		{"GET", "/badz", "", func(t *testing.T, req *http.Request, resp *http.Response) {
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		}},
		{"GET", "/healthz", "", func(t *testing.T, req *http.Request, resp *http.Response) {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}},
		{"GET", "/readyz", "", func(t *testing.T, req *http.Request, resp *http.Response) {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}},
		{"GET", "/version", "", func(t *testing.T, req *http.Request, resp *http.Response) {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			details := build.Details{}
			err := json.NewDecoder(resp.Body).Decode(&details)
			require.NoError(t, err)
			assert.EqualValues(t, build.Data(), details)
		}},

		// Default sumdb is sum.golang.org
		{"GET", "/sumdb/sum.golang.org/supported", "", func(t *testing.T, req *http.Request, resp *http.Response) {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}},
		{"GET", "/sumdb/sum.rust-lang.org/supported", "", func(t *testing.T, req *http.Request, resp *http.Response) {
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		}},
		{"GET", "/sumdb/sum.golang.org/lookup/github.com/gomods/athens", "", func(t *testing.T, req *http.Request, resp *http.Response) {
			assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		}},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(
			tc.method,
			baseURL+tc.path,
			strings.NewReader(tc.body),
		)
		t.Run(req.RequestURI, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			tc.test(t, req, w.Result())
		})
	}
}

func TestGoDownloadRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/dl/" && r.URL.Query().Get("mode") == "json":
			_, _ = io.WriteString(w, `[{"version":"go1.27.1","stable":true,"files":[{"filename":"go1.27.1.linux-amd64.tar.gz","version":"go1.27.1","os":"linux","arch":"amd64","kind":"archive","size":7,"sha256":"db4b4d0d1cb480bf9aeea253771c00febe627f236765fa37d6a5614f079a3aa0"}]}]`)
		case r.URL.Path == "/dl/go1.27.1.linux-amd64.tar.gz":
			_, _ = io.WriteString(w, "tarball")
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	r := mux.NewRouter()
	s, err := mem.NewStorage()
	require.NoError(t, err)
	c, err := config.Load("")
	require.NoError(t, err)
	c.NoSumPatterns = []string{"*"}
	c.PathPrefix = "/prefix"
	c.GoDownloadURL = upstream.URL + "/dl"

	subRouter := r.PathPrefix(c.PathPrefix).Subrouter()
	require.NoError(t, addProxyRoutes(t.Context(), subRouter, s, log.NoOpLogger(), c))

	for path, want := range map[string]string{
		"/prefix/dl/?mode=json&include=all":      `[{"version":"go1.27.1","stable":true,"files":[{"filename":"go1.27.1.linux-amd64.tar.gz","version":"go1.27.1","os":"linux","arch":"amd64","kind":"archive","size":7,"sha256":"db4b4d0d1cb480bf9aeea253771c00febe627f236765fa37d6a5614f079a3aa0"}]}]`,
		"/prefix/dl/go1.27.1.linux-amd64.tar.gz": "tarball",
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		assert.Equal(t, http.StatusOK, w.Code, path)
		if len(want) > 0 && want[0] == '[' {
			assert.JSONEq(t, want, w.Body.String(), path)
		} else {
			assert.Equal(t, want, w.Body.String(), path)
		}
	}

	// A module named "dl" must still route to the download protocol, not the
	// toolchain proxy.
	for _, path := range []string{"/prefix/dl/@v/list", "/prefix/dl/@latest", "/prefix/dl/@v/v1.0.0.zip"} {
		var match mux.RouteMatch
		require.True(t, r.Match(httptest.NewRequest(http.MethodGet, path, nil), &match), path)
		assert.NotEqual(t, goDownloadRouteName, match.Route.GetName(), path)
	}
}

func TestGoDownloadRoutesDisabledByDefault(t *testing.T) {
	r := mux.NewRouter()
	s, err := mem.NewStorage()
	require.NoError(t, err)
	c, err := config.Load("")
	require.NoError(t, err)
	c.NoSumPatterns = []string{"*"}
	require.NoError(t, addProxyRoutes(t.Context(), r, s, log.NoOpLogger(), c))

	for _, path := range []string{"/dl/go1.27.1.linux-amd64.tar.gz", "/dl/?mode=json&include=all"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, path)
		assert.Contains(t, w.Body.String(), "ATHENS_GO_DOWNLOAD_URL", path)
	}
}

func TestOfflineGoDownloadsUseAuthentication(t *testing.T) {
	backend, err := mem.NewStorage()
	require.NoError(t, err)
	c, err := config.Load("")
	require.NoError(t, err)
	c.GoDownloadEnabled = true
	c.NetworkMode = "offline"
	c.PathPrefix = "/prefix"
	router := mux.NewRouter()
	router.Use(basicAuth("release-user", "release-password"))
	require.NoError(t, addGoDownloadRoutes(t.Context(), router.PathPrefix(c.PathPrefix).Subrouter(), backend, c))
	for _, authenticated := range []bool{false, true} {
		req := httptest.NewRequest(http.MethodGet, "/prefix/dl/?mode=json&include=all", nil)
		if authenticated {
			req.SetBasicAuth("release-user", "release-password")
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		if authenticated {
			require.Equal(t, http.StatusOK, response.Code)
			require.JSONEq(t, "[]", response.Body.String())
		} else {
			require.Equal(t, http.StatusUnauthorized, response.Code)
		}
	}
}
