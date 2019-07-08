package download

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gorilla/mux"
)

func TestRedirect(t *testing.T) {
	r := mux.NewRouter()
	RegisterHandlers(r, &HandlerOpts{
		Protocol: &mockProtocol{},
		Logger:   log.NoOpLogger(),
		DownloadFile: &mode.DownloadFile{
			Mode:        mode.Redirect,
			DownloadURL: "https://gomods.io",
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
		expectedRedirect := "https://gomods.io" + path
		givenRedirect := w.HeaderMap.Get("location")
		if expectedRedirect != givenRedirect {
			t.Fatalf("expected the handler to redirect to %q but got %q", expectedRedirect, givenRedirect)
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

func (mp *mockProtocol) Zip(ctx context.Context, mod, ver string) (io.ReadCloser, error) {
	const op errors.Op = "mockProtocol.Zip"
	return nil, errors.E(op, "not found", errors.KindRedirect)
}
