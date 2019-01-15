package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
)

// CacheControl takes a string and makes a header value to the key Cache-Control.
// This is so you can set some sane cache defaults to certain endpoints.
func CacheControl(cacheHeaderValue string) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", cacheHeaderValue)
			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(f)
	}
}
