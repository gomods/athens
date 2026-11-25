package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestCacheControl(t *testing.T) {
	const expected = "public, max-age=31536000, immutable"
	const noCache = CacheControlNoCache

	t.Run("leaves header unset when empty", func(t *testing.T) {
		ch := CacheControl("")
		handler := ch(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/test", nil)
		handler.ServeHTTP(w, r)
		if got := w.Result().Header.Get("Cache-Control"); got != "" {
			t.Fatalf("expected unset Cache-Control, got %q", got)
		}
	})

	t.Run("sets header on 200 for canonical versions", func(t *testing.T) {
		for _, ver := range []string{
			"v1.0.0",
			"v2.0.0+incompatible",
			"v0.0.0-20210101000000-abcdef123456",
			"v1.0.0-!r!c1",
		} {
			t.Run(ver, func(t *testing.T) {
				r := mux.NewRouter()
				r.Handle("/{module:.+}/@v/{version}.info", CacheControl(expected)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})))
				req := httptest.NewRequest(http.MethodGet, "/github.com/foo/bar/@v/"+ver+".info", nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				if got := w.Result().Header.Get("Cache-Control"); got != expected {
					t.Fatalf("expected Cache-Control %q for %s, got %q", expected, ver, got)
				}
			})
		}
	})

	t.Run("sets no-cache on non-canonical versions", func(t *testing.T) {
		for _, ver := range []string{"main", "v1.0", "v1.0.0+build"} {
			t.Run(ver, func(t *testing.T) {
				r := mux.NewRouter()
				r.Handle("/{module:.+}/@v/{version}.info", CacheControl(expected)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})))
				req := httptest.NewRequest(http.MethodGet, "/github.com/foo/bar/@v/"+ver+".info", nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				if got := w.Result().Header.Get("Cache-Control"); got != noCache {
					t.Fatalf("expected Cache-Control %q for %s, got %q", noCache, ver, got)
				}
			})
		}
	})

	t.Run("sets no-cache on non-200 responses", func(t *testing.T) {
		for _, status := range []int{
			http.StatusMovedPermanently,
			http.StatusBadRequest,
			http.StatusNotFound,
			http.StatusGone,
			http.StatusInternalServerError,
		} {
			r := mux.NewRouter()
			r.Handle("/{module:.+}/@v/{version}.info", CacheControl(expected)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
			})))
			req := httptest.NewRequest(http.MethodGet, "/github.com/foo/bar/@v/v1.0.0.info", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if got := w.Result().Header.Get("Cache-Control"); got != noCache {
				t.Fatalf("expected Cache-Control %q for status %d, got %q", noCache, status, got)
			}
		}
	})

	t.Run("implicit 200 Write prevents subsequent WriteHeader from overwriting Cache-Control", func(t *testing.T) {
		r := mux.NewRouter()
		r.Handle("/{module:.+}/@v/{version}.info", CacheControl(expected)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok"))
			w.WriteHeader(http.StatusInternalServerError)
		})))
		req := httptest.NewRequest(http.MethodGet, "/github.com/foo/bar/@v/v1.0.0.info", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if got := w.Result().Header.Get("Cache-Control"); got != expected {
			t.Fatalf("expected Cache-Control %q, got %q", expected, got)
		}
	})
	t.Run("sets header on non-version routes on 200", func(t *testing.T) {
		handler := CacheControl(expected)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		handler.ServeHTTP(w, req)
		if got := w.Result().Header.Get("Cache-Control"); got != expected {
			t.Fatalf("expected Cache-Control %q, got %q", expected, got)
		}
	})
}

func TestCacheControlUnwrap(t *testing.T) {
	handler := CacheControl("public, max-age=60")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := http.NewResponseController(w).Flush(); err != nil {
			t.Fatalf("expected ResponseController.Flush to succeed via Unwrap: %v", err)
		}
	}))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	handler.ServeHTTP(w, req)
	if !w.Flushed {
		t.Fatal("expected underlying ResponseRecorder to be flushed")
	}
}
