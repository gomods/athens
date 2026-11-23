package middleware

import (
	"net/http"

	"github.com/gomods/athens/pkg/paths"
	"github.com/gorilla/mux"
	"golang.org/x/mod/module"
)

// CacheControlNoCache is the Cache-Control value for responses that must not
// be cached.
const CacheControlNoCache = "no-cache, no-store, must-revalidate"

// CacheControl returns a middleware that sets the Cache-Control header to
// cacheHeaderValue, so you can set cache defaults on certain endpoints.
//
// The value is only sent where caching it is safe:
//   - on HTTP 200 responses. Any other status (404, a redirect, an error) gets
//     CacheControlNoCache, so a caching proxy never keeps a "not found" for as
//     long as it would keep a module file.
//   - when the request has no module version, or the version is canonical
//     (v1.2.3, a pseudo-version, v2.0.0+incompatible). A branch, a revision or
//     a short version such as "main" or "v1.2" is resolved to something that
//     can change, so it gets CacheControlNoCache.
//
// An empty cacheHeaderValue disables the middleware.
func CacheControl(cacheHeaderValue string) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		if cacheHeaderValue == "" {
			return h
		}
		f := func(w http.ResponseWriter, r *http.Request) {
			value := cacheHeaderValue
			if ver, err := paths.GetVersion(r); err == nil && module.CanonicalVersion(ver) != ver {
				value = CacheControlNoCache
			}
			w.Header().Set("Cache-Control", value)
			h.ServeHTTP(&cacheControlWriter{ResponseWriter: w}, r)
		}

		return http.HandlerFunc(f)
	}
}

// cacheControlWriter replaces the Cache-Control header with CacheControlNoCache
// when the handler answers with anything but 200. An implicit 200 (a Write with
// no WriteHeader) keeps the value that was set up front.
type cacheControlWriter struct {
	http.ResponseWriter
}

func (w *cacheControlWriter) WriteHeader(statusCode int) {
	if statusCode != http.StatusOK {
		w.Header().Set("Cache-Control", CacheControlNoCache)
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (w *cacheControlWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
