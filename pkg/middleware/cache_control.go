package middleware

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gorilla/mux"
)

// CacheControl takes a string and makes a header value to the key Cache-Control.
// This is so you can set some sane cache defaults to certain endpoints.
func CacheControl(cacheHeaderValue string) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			c.Response().Header().Set("Cache-Control", cacheHeaderValue)
			return next(c)
		}
	}
}

// CacheControlV2 is an implementation of CacheControl as a mux.MiddlewareFunc
func CacheControlV2(cacheHeaderValue string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", cacheHeaderValue)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
