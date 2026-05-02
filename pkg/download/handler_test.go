package download

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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


func TestInfoHandlerGoneIncludesErrorMessage(t *testing.T) {
	r := mux.NewRouter()
	RegisterHandlers(r, &HandlerOpts{
		Protocol: &mockGoneInfoProtocol{},
		Logger:   log.NoOpLogger(),
		DownloadFile: &mode.DownloadFile{
			Mode: mode.Sync,
		},
	})

	req := httptest.NewRequest("GET", "/github.com/uber/h3-go/@v/v3.0.2.info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusGone {
		t.Fatalf("expected status %d but got %d", http.StatusGone, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "major version must be compatible") {
		t.Fatalf("expected descriptive semver error in body, got %q", body)
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

type mockGoneInfoProtocol struct {
	Protocol
}

func (mp *mockGoneInfoProtocol) Info(ctx context.Context, mod, ver string) ([]byte, error) {
	const op errors.Op = "mockGoneInfoProtocol.Info"
	return nil, errors.E(op, "invalid version: module contains a go.mod file, so major version must be compatible: should be v0 or v1, not v3", errors.KindGone)
}
