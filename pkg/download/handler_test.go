package download

import (
	"context"
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

func TestContentCacheControl(t *testing.T) {
	contentPaths := [...]string{
		"/github.com/gomods/athens/@v/v0.4.0.info",
		"/github.com/gomods/athens/@v/v0.4.0.mod",
		"/github.com/gomods/athens/@v/v0.4.0.zip",
	}

	t.Run("sets the header when configured", func(t *testing.T) {
		const cacheControl = "public, max-age=300"
		r := mux.NewRouter()
		RegisterHandlers(r, &HandlerOpts{
			Protocol:            &mockProtocol{},
			Logger:              log.NoOpLogger(),
			DownloadFile:        &mode.DownloadFile{Mode: mode.Redirect, DownloadURL: "https://gomods.io"},
			ContentCacheControl: cacheControl,
		})
		for _, path := range contentPaths {
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if got := w.Header().Get("Cache-Control"); got != cacheControl {
				t.Fatalf("expected Cache-Control %q on %s but got %q", cacheControl, path, got)
			}
		}
	})

	t.Run("leaves the header off by default", func(t *testing.T) {
		r := mux.NewRouter()
		RegisterHandlers(r, &HandlerOpts{
			Protocol:     &mockProtocol{},
			Logger:       log.NoOpLogger(),
			DownloadFile: &mode.DownloadFile{Mode: mode.Redirect, DownloadURL: "https://gomods.io"},
		})
		for _, path := range contentPaths {
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if got := w.Header().Get("Cache-Control"); got != "" {
				t.Fatalf("expected no Cache-Control on %s but got %q", path, got)
			}
		}
	})
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
