package external

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/stretchr/testify/require"
)

func TestToolchainCapabilityNegotiation(t *testing.T) {
	legacy := httptest.NewServer(http.NotFoundHandler())
	defer legacy.Close()
	client := NewClient(legacy.URL, nil).(*service)
	require.ErrorContains(t, client.CheckToolchainStorage(t.Context()), "must support")
	backend, err := mem.NewStorage()
	require.NoError(t, err)
	server := httptest.NewServer(NewServer(backend))
	defer server.Close()
	client = NewClient(server.URL, nil).(*service)
	require.NoError(t, client.CheckToolchainStorage(t.Context()))
}

func TestMalformedReleaseWritesAreRejected(t *testing.T) {
	backend, err := mem.NewStorage()
	require.NoError(t, err)
	handler := NewServer(backend)
	for _, target := range []string{"/toolchains/v1/archive/go1.25.1.linux-amd64.tar.gz?source=official", "/toolchains/v1/releases?source=official"} {
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader("invalid JSON"))
		req.Header.Set(archiveInfoHeader, "invalid base64!")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
	}
}
