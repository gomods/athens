package middleware

import (
	"net/http"

	"github.com/gomods/athens/pkg/auth"
)

type authkey struct{}

// WithAuth inserts the Authorization header
// into the request context
func WithAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if ok {
			ctx := auth.SetAuthInContext(r.Context(), auth.BasicAuth{User: user, Password: password})
			r = r.WithContext(ctx)
		}
		h.ServeHTTP(w, r)
	})
}
