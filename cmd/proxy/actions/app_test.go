package actions

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppReturnsCleanup(t *testing.T) {
	l := log.NoOpLogger()
	c, err := config.Load("")
	require.NoError(t, err)

	handler, cleanup, err := App(l, c)
	require.NoError(t, err)
	assert.NotNil(t, handler)
	assert.NotNil(t, cleanup)

	// cleanup should be safe to call without panic.
	assert.NotPanics(t, cleanup)
}

func TestAppReturnsCleanupWithExporters(t *testing.T) {
	l := log.NoOpLogger()
	c, err := config.Load("")
	require.NoError(t, err)
	// Exercise the real exporter registration path: OTLP traces (the gRPC
	// exporter dials lazily, so no collector is needed) and Prometheus metrics.
	c.TraceExporter = "otlp"
	c.StatsExporter = "prometheus"

	handler, cleanup, err := App(l, c)
	require.NoError(t, err)
	assert.NotNil(t, handler)
	assert.NotNil(t, cleanup)

	// cleanup shuts down the trace and metric providers and must be safe to call.
	assert.NotPanics(t, cleanup)

	// Calling cleanup a second time should also be safe (idempotency).
	assert.NotPanics(t, cleanup)
}

func TestAppUpstreamRedirectsOmitPathPrefix(t *testing.T) {
	for _, prefix := range []string{"/prefix", "/prefix/"} {
		t.Run(prefix, func(t *testing.T) {
			filterFile := filepath.Join(t.TempDir(), "filter.conf")
			require.NoError(t, os.WriteFile(filterFile, []byte("D example.com/direct\n"), 0o600))

			c, err := config.Load("")
			require.NoError(t, err)
			c.PathPrefix = prefix
			c.DownloadMode = mode.Redirect
			c.DownloadURL = "https://proxy.example.com"
			c.FilterFile = filterFile
			c.GlobalEndpoint = "https://upstream.example.com"

			handler, cleanup, err := App(log.NoOpLogger(), c)
			require.NoError(t, err)
			t.Cleanup(cleanup)

			for _, tc := range []struct {
				path     string
				code     int
				location string
			}{
				// Download mode redirects.
				{"/prefix/example.com/redirect/@v/v1.0.0.info", http.StatusMovedPermanently, "https://proxy.example.com/example.com/redirect/@v/v1.0.0.info"},
				{"/prefix/example.com/redirect/@v/v1.0.0.mod", http.StatusMovedPermanently, "https://proxy.example.com/example.com/redirect/@v/v1.0.0.mod"},
				{"/prefix/example.com/redirect/@v/v1.0.0.zip", http.StatusMovedPermanently, "https://proxy.example.com/example.com/redirect/@v/v1.0.0.zip"},
				// Filter file "D" (direct) redirect.
				{"/prefix/example.com/direct/@v/list", http.StatusSeeOther, "https://upstream.example.com/example.com/direct/@v/list"},
			} {
				req := httptest.NewRequest(http.MethodGet, tc.path, nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				assert.Equal(t, tc.code, w.Code, tc.path)
				assert.Equal(t, tc.location, w.Header().Get("Location"), tc.path)
			}
		})
	}
}
