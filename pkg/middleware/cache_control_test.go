package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestCacheControl(t *testing.T) {
	const value = "public, max-age=31536000, immutable"

	status := func(code int) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(code) }
	}
	redirect := func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/x", http.StatusMovedPermanently) }
	implicit := func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("body")) }

	tests := []struct {
		name    string
		value   string
		path    string
		handler http.HandlerFunc
		want    string // "" means the header must be absent
	}{
		{"canonical release", value, "/m/@v/v1.2.3.info", status(200), value},
		{"pseudo-version", value, "/m/@v/v0.0.0-20200101000000-abcdef123456.info", status(200), value},
		{"prerelease", value, "/m/@v/v1.2.3-rc.1.info", status(200), value},
		{"incompatible is immutable", value, "/m/@v/v2.0.0+incompatible.info", status(200), value},
		{"case-encoded version is decoded", value, "/m/@v/v1.0.0-!r!c1.info", status(200), value},
		{"implicit 200", value, "/m/@v/v1.2.3.info", implicit, value},
		{"branch name", value, "/m/@v/main.info", status(200), CacheControlNoCache},
		{"master", value, "/m/@v/master.info", status(200), CacheControlNoCache},
		{"short version", value, "/m/@v/v1.2.info", status(200), CacheControlNoCache},
		{"major only", value, "/m/@v/v1.info", status(200), CacheControlNoCache},
		{"not found", value, "/m/@v/v1.2.3.info", status(404), CacheControlNoCache},
		{"gone", value, "/m/@v/v1.2.3.info", status(410), CacheControlNoCache},
		{"server error", value, "/m/@v/v1.2.3.info", status(500), CacheControlNoCache},
		{"bad request", value, "/m/@v/v1.2.3.info", status(400), CacheControlNoCache},
		{"redirect", value, "/m/@v/v1.2.3.info", redirect, CacheControlNoCache},
		{"empty value is a no-op", "", "/m/@v/v1.2.3.info", status(200), ""},
		{"no version in the route", value, "/m/@v/list", status(200), value},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := mux.NewRouter()
			mw := CacheControl(tc.value)
			r.Handle("/{module:.+}/@v/{version}.info", mw(tc.handler))
			r.Handle("/{module:.+}/@v/list", mw(tc.handler))

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))

			got := rec.Header().Values("Cache-Control")
			if tc.want == "" {
				if len(got) != 0 {
					t.Fatalf("expected no Cache-Control header, got %q", got)
				}
				return
			}
			if len(got) != 1 || got[0] != tc.want {
				t.Fatalf("Cache-Control = %q, want %q", got, tc.want)
			}
		})
	}
}

// Handlers and middlewares further down rely on http.ResponseController to
// reach the real writer.
func TestCacheControlUnwrap(t *testing.T) {
	h := CacheControl("public")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := http.NewResponseController(w).Flush(); err != nil {
			t.Errorf("Flush through the wrapper: %v", err)
		}
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if !rec.Flushed {
		t.Fatal("expected the flush to reach the underlying writer")
	}
}
