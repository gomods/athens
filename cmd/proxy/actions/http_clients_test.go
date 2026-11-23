package actions

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mw "github.com/gomods/athens/pkg/middleware"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

const testTimeout = 150 * time.Millisecond

// hangingServer never answers until the test is over.
func hangingServer(t *testing.T) *httptest.Server {
	t.Helper()
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		case <-time.After(30 * time.Second):
		}
	}))
	t.Cleanup(func() {
		close(release)
		srv.Close()
	})
	return srv
}

func TestValidatorClientTimesOutOnHungWebhook(t *testing.T) {
	srv := hangingServer(t)
	validator, _ := newHTTPClients(testTimeout)
	require.Equal(t, testTimeout, validator.Timeout)

	start := time.Now()
	resp, err := validator.Get(srv.URL)
	if resp != nil {
		resp.Body.Close()
	}
	require.Error(t, err)
	require.Less(t, time.Since(start), 5*time.Second)
}

func TestValidationMiddlewareAnswers500OnHungWebhook(t *testing.T) {
	srv := hangingServer(t)
	validator, _ := newHTTPClients(testTimeout)

	r := mux.NewRouter()
	r.Use(mw.NewValidationMiddleware(validator, srv.URL))
	r.HandleFunc("/{module:.+}/@v/{version}.info", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	start := time.Now()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/github.com/a/b/@v/v1.0.0.info", nil))
	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Less(t, time.Since(start), 5*time.Second)
}

func TestStorageClientConfig(t *testing.T) {
	_, storage := newHTTPClients(testTimeout)
	require.Zero(t, storage.Timeout, "Client.Timeout would cut off streamed bodies")
	require.NotNil(t, storage.Transport)

	tr := newStorageTransport(testTimeout)
	require.NotNil(t, tr)
	require.Equal(t, testTimeout, tr.ResponseHeaderTimeout)

	// A non-positive timeout must not set a header timeout.
	require.Zero(t, newStorageTransport(0).ResponseHeaderTimeout)

	// The shared http.DefaultTransport must not be mutated.
	dt, ok := http.DefaultTransport.(*http.Transport)
	require.True(t, ok)
	require.Zero(t, dt.ResponseHeaderTimeout)
}

func TestStorageClientTimesOutWaitingForHeaders(t *testing.T) {
	srv := hangingServer(t)
	_, storage := newHTTPClients(testTimeout)

	start := time.Now()
	resp, err := storage.Get(srv.URL)
	if resp != nil {
		resp.Body.Close()
	}
	require.Error(t, err)
	require.Less(t, time.Since(start), 5*time.Second)
}

func TestStorageClientDoesNotCutStreamedBody(t *testing.T) {
	const chunks = 8
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		f := w.(http.Flusher)
		f.Flush() // headers go out immediately
		for i := 0; i < chunks; i++ {
			time.Sleep(testTimeout / 2)
			_, _ = io.WriteString(w, "chunk")
			f.Flush()
		}
	}))
	t.Cleanup(srv.Close)

	_, storage := newHTTPClients(testTimeout)
	start := time.Now()
	resp, err := storage.Get(srv.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, strings.Repeat("chunk", chunks), string(body))
	require.Greater(t, time.Since(start), testTimeout, "body must outlast the header timeout")
}
