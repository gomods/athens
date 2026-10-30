package middleware

import (
	"net/http"

	"github.com/gomods/athens/pkg/paths"
	"github.com/gorilla/mux"
	"golang.org/x/mod/semver"
)

// CacheControl takes a string and makes a header value to the key Cache-Control.
// The configured cacheHeaderValue is only applied to HTTP 200 responses.
// If the request contains a module version parameter, it must be a canonical semantic version;
// non-canonical versions (e.g. branch names like /@v/main.info) receive "no-cache, no-store, must-revalidate".
// Non-200 responses (e.g. 404 Not Found or errors) always receive "no-cache, no-store, must-revalidate".
func CacheControl(cacheHeaderValue string) mux.MiddlewareFunc {
	if cacheHeaderValue == "" {
		return func(h http.Handler) http.Handler {
			return h
		}
	}

	return func(h http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {
			isCanonical := true
			ver, err := paths.GetVersion(r)
			if err == nil {
				isCanonical = semver.IsValid(ver) && semver.Canonical(ver) == ver
			}

			if isCanonical {
				w.Header().Set("Cache-Control", cacheHeaderValue)
			} else {
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			}
			rw := &cacheControlResponseWriter{
				ResponseWriter: w,
				wroteHeader:    false,
			}
			h.ServeHTTP(rw, r)
		}

		return http.HandlerFunc(f)
	}
}

type cacheControlResponseWriter struct {
	http.ResponseWriter

	wroteHeader bool
}

func (w *cacheControlResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	if statusCode != http.StatusOK {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	w.ResponseWriter.WriteHeader(statusCode)
}
