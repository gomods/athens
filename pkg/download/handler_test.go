package download

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/middleware"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gorilla/mux"
)

func TestRedirect(t *testing.T) {
	for _, url := range []string{"https://gomods.io", "https://internal.domain/repository/gonexus"} {
		r := mux.NewRouter()
		RegisterHandlers(r, &HandlerOpts{
			Protocol: &mockProtocol{},
			Logger:   log.NoOpLogger(),
			DownloadFile: &mode.DownloadFile{
				Mode:        mode.Redirect,
				DownloadURL: url,
			},
		})
		for _, path := range [...]string{
			"/github.com/gomods/athens/@v/v0.4.0.info",
			"/github.com/gomods/athens/@v/v0.4.0.mod",
			"/github.com/gomods/athens/@v/v0.4.0.zip",
		} {
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusMovedPermanently {
				t.Fatalf("expected a redirect status (301) but got %v", w.Code)
			}
			expectedRedirect := url + path
			givenRedirect := w.HeaderMap.Get("location")
			if expectedRedirect != givenRedirect {
				t.Fatalf("expected the handler to redirect to %q but got %q", expectedRedirect, givenRedirect)
			}
		}
	}
}

type mockProtocol struct {
	Protocol
}

func (mp *mockProtocol) Info(ctx context.Context, mod, ver string) ([]byte, error) {
	const op errors.Op = "mockProtocol.Info"
	return nil, errors.E(op, "not found", errors.KindRedirect)
}

func (mp *mockProtocol) GoMod(ctx context.Context, mod, ver string) ([]byte, error) {
	const op errors.Op = "mockProtocol.GoMod"
	return nil, errors.E(op, "not found", errors.KindRedirect)
}

func (mp *mockProtocol) Zip(ctx context.Context, mod, ver string) (storage.SizeReadCloser, error) {
	const op errors.Op = "mockProtocol.Zip"
	return nil, errors.E(op, "not found", errors.KindRedirect)
}

// successProtocol serves every module version successfully.
type successProtocol struct {
	Protocol
}

func (successProtocol) List(context.Context, string) ([]string, error) {
	return []string{"v0.4.0"}, nil
}

func (successProtocol) Info(_ context.Context, _, ver string) ([]byte, error) {
	return []byte(`{"Version":"` + ver + `"}`), nil
}

func (successProtocol) Latest(context.Context, string) (*storage.RevInfo, error) {
	return &storage.RevInfo{Version: "v0.4.0"}, nil
}

func (successProtocol) GoMod(_ context.Context, mod, _ string) ([]byte, error) {
	return []byte("module " + mod), nil
}

func (successProtocol) Zip(context.Context, string, string) (storage.SizeReadCloser, error) {
	return storage.NewSizer(io.NopCloser(strings.NewReader("zip")), 3), nil
}

// notFoundProtocol finds nothing.
type notFoundProtocol struct {
	Protocol
}

func (notFoundProtocol) Info(context.Context, string, string) ([]byte, error) {
	return nil, errors.E("notFoundProtocol.Info", "not found", errors.KindNotFound)
}

func (notFoundProtocol) GoMod(context.Context, string, string) ([]byte, error) {
	return nil, errors.E("notFoundProtocol.GoMod", "not found", errors.KindNotFound)
}

func (notFoundProtocol) Zip(context.Context, string, string) (storage.SizeReadCloser, error) {
	return nil, errors.E("notFoundProtocol.Zip", "not found", errors.KindNotFound)
}

func TestCacheControl(t *testing.T) {
	const cacheControl = "public, max-age=31536000, immutable"
	syncMode := &mode.DownloadFile{Mode: mode.Sync}
	redirectMode := &mode.DownloadFile{Mode: mode.Redirect, DownloadURL: "https://gomods.io"}

	const base = "/github.com/gomods/athens/@v/"
	canonical := []string{base + "v0.4.0.info", base + "v0.4.0.mod", base + "v0.4.0.zip"}

	tests := []struct {
		name         string
		protocol     Protocol
		df           *mode.DownloadFile
		cacheControl string
		method       string
		paths        []string
		wantStatus   int
		want         string // "" means the header must be absent
	}{
		{"canonical versions get the configured value", successProtocol{}, syncMode, cacheControl, http.MethodGet, canonical, http.StatusOK, cacheControl},
		{"HEAD on a zip gets the configured value", successProtocol{}, syncMode, cacheControl, http.MethodHead, []string{base + "v0.4.0.zip"}, http.StatusOK, cacheControl},
		{"+incompatible is canonical", successProtocol{}, syncMode, cacheControl, http.MethodGet, []string{base + "v2.0.0+incompatible.info"}, http.StatusOK, cacheControl},
		{"a branch is never cached", successProtocol{}, syncMode, cacheControl, http.MethodGet, []string{base + "main.info"}, http.StatusOK, middleware.CacheControlNoCache},
		{"not found is never cached", notFoundProtocol{}, syncMode, cacheControl, http.MethodGet, canonical, http.StatusNotFound, middleware.CacheControlNoCache},
		{"a redirect is never cached", &mockProtocol{}, redirectMode, cacheControl, http.MethodGet, canonical, http.StatusMovedPermanently, middleware.CacheControlNoCache},
		{"list and latest stay no-cache", successProtocol{}, syncMode, cacheControl, http.MethodGet, []string{"/github.com/gomods/athens/@v/list", "/github.com/gomods/athens/@latest"}, http.StatusOK, middleware.CacheControlNoCache},
		{"unset by default", successProtocol{}, syncMode, "", http.MethodGet, canonical, http.StatusOK, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := mux.NewRouter()
			RegisterHandlers(r, &HandlerOpts{
				Protocol:     tc.protocol,
				Logger:       log.NoOpLogger(),
				DownloadFile: tc.df,
				CacheControl: tc.cacheControl,
			})
			for _, path := range tc.paths {
				w := httptest.NewRecorder()
				r.ServeHTTP(w, httptest.NewRequest(tc.method, path, nil))
				if w.Code != tc.wantStatus {
					t.Fatalf("%s %s: status = %d, want %d", tc.method, path, w.Code, tc.wantStatus)
				}
				if got := w.Header().Get("Cache-Control"); got != tc.want {
					t.Fatalf("%s %s: Cache-Control = %q, want %q", tc.method, path, got, tc.want)
				}
			}
		})
	}
}

// countingWriter records how many times the handler writes a body.
type countingWriter struct {
	http.ResponseWriter
	writes int
}

func (w *countingWriter) Write(b []byte) (int, error) {
	w.writes++
	return w.ResponseWriter.Write(b)
}

// An error status is the whole response: the info handler must not go on to
// write a (nil) body after it.
func TestInfoHandlerWritesNoBodyOnError(t *testing.T) {
	r := mux.NewRouter()
	RegisterHandlers(r, &HandlerOpts{
		Protocol:     notFoundProtocol{},
		Logger:       log.NoOpLogger(),
		DownloadFile: &mode.DownloadFile{Mode: mode.Sync},
	})
	w := &countingWriter{ResponseWriter: httptest.NewRecorder()}
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/github.com/gomods/athens/@v/v0.4.0.info", nil))
	if w.writes != 0 {
		t.Fatalf("expected no body write after the error status, got %d", w.writes)
	}
}
