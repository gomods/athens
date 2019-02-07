package middleware

import (
	"net/http"
)

// ContentType writes writes an application/json
// Content-Type header.
func ContentType(h http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(f)
}
