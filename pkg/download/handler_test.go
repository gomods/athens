package download

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
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

type mockZip struct {
	io.ReadCloser
	size int64
}

func (m *mockZip) Size() int64 {
	return m.size
}

type mockSuccessProtocol struct {
	Protocol
}

func (mp *mockSuccessProtocol) Info(ctx context.Context, mod, ver string) ([]byte, error) {
	return []byte(`{"Version":"` + ver + `"}`), nil
}

func (mp *mockSuccessProtocol) GoMod(ctx context.Context, mod, ver string) ([]byte, error) {
	return []byte("module " + mod), nil
}

func (mp *mockSuccessProtocol) Zip(ctx context.Context, mod, ver string) (storage.SizeReadCloser, error) {
	content := []byte("zip content")
	return &mockZip{ReadCloser: io.NopCloser(bytes.NewReader(content)), size: int64(len(content))}, nil
}

type mockNotFoundProtocol struct {
	Protocol
}

func (mp *mockNotFoundProtocol) Info(ctx context.Context, mod, ver string) ([]byte, error) {
	return nil, errors.E("mockNotFoundProtocol.Info", "not found", errors.KindNotFound)
}

func (mp *mockNotFoundProtocol) GoMod(ctx context.Context, mod, ver string) ([]byte, error) {
	return nil, errors.E("mockNotFoundProtocol.GoMod", "not found", errors.KindNotFound)
}

func (mp *mockNotFoundProtocol) Zip(ctx context.Context, mod, ver string) (storage.SizeReadCloser, error) {
	return nil, errors.E("mockNotFoundProtocol.Zip", "not found", errors.KindNotFound)
}

func TestCacheControl(t *testing.T) {
	const cacheControl = "public, max-age=31536000, immutable"
	canonicalFilePaths := [...]string{
		"/github.com/gomods/athens/@v/v0.4.0.info",
		"/github.com/gomods/athens/@v/v0.4.0.mod",
		"/github.com/gomods/athens/@v/v0.4.0.zip",
	}

	t.Run("sets the header on 200 responses for canonical module versions when configured", func(t *testing.T) {
		r := mux.NewRouter()
		RegisterHandlers(r, &HandlerOpts{
			Protocol:     &mockSuccessProtocol{},
			Logger:       log.NoOpLogger(),
			DownloadFile: &mode.DownloadFile{Mode: mode.Sync},
			CacheControl: cacheControl,
		})
		for _, path := range canonicalFilePaths {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200 for %s but got %d", path, w.Code)
			}
			if got := w.Header().Get("Cache-Control"); got != cacheControl {
				t.Fatalf("expected Cache-Control %q for %s but got %q", cacheControl, path, got)
			}
		}

		// Also test HEAD request on .zip
		headReq := httptest.NewRequest(http.MethodHead, "/github.com/gomods/athens/@v/v0.4.0.zip", nil)
		headW := httptest.NewRecorder()
		r.ServeHTTP(headW, headReq)
		if headW.Code != http.StatusOK {
			t.Fatalf("expected status 200 for HEAD but got %d", headW.Code)
		}
		if got := headW.Header().Get("Cache-Control"); got != cacheControl {
			t.Fatalf("expected Cache-Control %q for HEAD but got %q", cacheControl, got)
		}
	})

	t.Run("sets no-cache on non-canonical versions", func(t *testing.T) {
		r := mux.NewRouter()
		RegisterHandlers(r, &HandlerOpts{
			Protocol:     &mockSuccessProtocol{},
			Logger:       log.NoOpLogger(),
			DownloadFile: &mode.DownloadFile{Mode: mode.Sync},
			CacheControl: cacheControl,
		})
		req := httptest.NewRequest(http.MethodGet, "/github.com/gomods/athens/@v/main.info", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 but got %d", w.Code)
		}
		const expected = "no-cache, no-store, must-revalidate"
		if got := w.Header().Get("Cache-Control"); got != expected {
			t.Fatalf("expected Cache-Control %q for main.info but got %q", expected, got)
		}
	})

	t.Run("does not set immutable header on error responses", func(t *testing.T) {
		r := mux.NewRouter()
		RegisterHandlers(r, &HandlerOpts{
			Protocol:     &mockNotFoundProtocol{},
			Logger:       log.NoOpLogger(),
			DownloadFile: &mode.DownloadFile{Mode: mode.Sync},
			CacheControl: cacheControl,
		})
		for _, path := range canonicalFilePaths {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusNotFound {
				t.Fatalf("expected status 404 for %s but got %d", path, w.Code)
			}
			if got := w.Header().Get("Cache-Control"); got == cacheControl {
				t.Fatalf("expected error response to not have Cache-Control %q for %s", cacheControl, path)
			}
		}
	})

	t.Run("does not set immutable header on redirect responses", func(t *testing.T) {
		r := mux.NewRouter()
		RegisterHandlers(r, &HandlerOpts{
			Protocol:     &mockProtocol{},
			Logger:       log.NoOpLogger(),
			DownloadFile: &mode.DownloadFile{Mode: mode.Redirect, DownloadURL: "https://gomods.io"},
			CacheControl: cacheControl,
		})
		for _, path := range canonicalFilePaths {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusMovedPermanently {
				t.Fatalf("expected status 301 for %s but got %d", path, w.Code)
			}
			if got := w.Header().Get("Cache-Control"); got == cacheControl {
				t.Fatalf("expected redirect response to not have Cache-Control %q for %s", cacheControl, path)
			}
		}
	})

	t.Run("leaves the header unset by default", func(t *testing.T) {
		r := mux.NewRouter()
		RegisterHandlers(r, &HandlerOpts{
			Protocol:     &mockSuccessProtocol{},
			Logger:       log.NoOpLogger(),
			DownloadFile: &mode.DownloadFile{Mode: mode.Sync},
		})
		for _, path := range canonicalFilePaths {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if got := w.Header().Get("Cache-Control"); got != "" {
				t.Fatalf("expected no Cache-Control header for %s but got %q", path, got)
			}
		}
	})
}
