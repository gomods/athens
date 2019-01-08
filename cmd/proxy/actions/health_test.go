package actions

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	wantStatusCode := http.StatusOK
	wantBody := ""

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthHandler(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, resp.StatusCode, wantStatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("ioutil.ReadAll error = %s; want nil", err)
	}
	require.Equal(t, string(body), wantBody)
}
