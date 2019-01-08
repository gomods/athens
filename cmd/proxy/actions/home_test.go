package actions

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProxyHomeHandler(t *testing.T) {
	wantStatusCode := http.StatusOK
	wantBody := `"Welcome to The Athens Proxy"`

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	proxyHomeHandler(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, resp.StatusCode, wantStatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, string(body), wantBody)
}
