package verify_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gomods/athens/pkg/verify"
	"github.com/stretchr/testify/require"
)

func TestSumDBOracle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/lookup/example.com/m@v1.0.0":
			fmt.Fprint(w, "12345\nexample.com/m v1.0.0 h1:GOOD=\nexample.com/m v1.0.0/go.mod h1:MOD=\n")
		case "/lookup/example.com/missing@v1.0.0":
			http.Error(w, "not found", http.StatusNotFound)
		case "/lookup/example.com/gone@v1.0.0":
			http.Error(w, "gone", http.StatusGone)
		case "/lookup/example.com/nozip@v1.0.0":
			fmt.Fprint(w, "12345\nexample.com/nozip v1.0.0/go.mod h1:MOD=\n")
		default:
			http.Error(w, "unexpected", http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	o := verify.NewSumDBOracle(srv.URL, srv.Client())

	h, err := o.CanonicalHash(context.Background(), "example.com/m", "v1.0.0")
	require.NoError(t, err)
	require.Equal(t, "h1:GOOD=", h)

	_, err = o.CanonicalHash(context.Background(), "example.com/missing", "v1.0.0")
	require.ErrorIs(t, err, verify.ErrNotInSumDB)

	_, err = o.CanonicalHash(context.Background(), "example.com/gone", "v1.0.0")
	require.ErrorIs(t, err, verify.ErrNotInSumDB)

	_, err = o.CanonicalHash(context.Background(), "example.com/nozip", "v1.0.0")
	require.ErrorIs(t, err, verify.ErrNotInSumDB)
}
