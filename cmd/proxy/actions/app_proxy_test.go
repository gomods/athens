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
	err = addProxyRoutes(subRouter, s, l, c)
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
