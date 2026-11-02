package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestCacheControl(t *testing.T) {
	const expected = "public, max-age=31536000, immutable"
	const noCache = "no-cache, no-store, must-revalidate"

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

	t.Run("sets header on 200 for canonical version", func(t *testing.T) {
		r := mux.NewRouter()
		r.Handle("/{module:.+}/@v/{version}.info", CacheControl(expected)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})))
		req := httptest.NewRequest(http.MethodGet, "/github.com/foo/bar/@v/v1.0.0.info", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if got := w.Result().Header.Get("Cache-Control"); got != expected {
			t.Fatalf("expected Cache-Control %q, got %q", expected, got)
		}
	})

	t.Run("sets no-cache on non-canonical version", func(t *testing.T) {
		r := mux.NewRouter()
		r.Handle("/{module:.+}/@v/{version}.info", CacheControl(expected)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})))
		req := httptest.NewRequest(http.MethodGet, "/github.com/foo/bar/@v/main.info", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if got := w.Result().Header.Get("Cache-Control"); got != noCache {
			t.Fatalf("expected Cache-Control %q, got %q", noCache, got)
		}
	})

	t.Run("sets no-cache on non-200 responses", func(t *testing.T) {
		r := mux.NewRouter()
		r.Handle("/{module:.+}/@v/{version}.info", CacheControl(expected)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})))
		req := httptest.NewRequest(http.MethodGet, "/github.com/foo/bar/@v/v1.0.0.info", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if got := w.Result().Header.Get("Cache-Control"); got != noCache {
			t.Fatalf("expected Cache-Control %q, got %q", noCache, got)
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
