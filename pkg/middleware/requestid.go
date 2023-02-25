package middleware

import (
	"net/http"

	"github.com/gomods/athens/pkg/requestid"
	"github.com/google/uuid"
)

// WithRequestID ensures a request id is in the
// request context by either the incoming header
// or creating a new one.
func WithRequestID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(requestid.HeaderKey)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx := requestid.SetInContext(r.Context(), requestID)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}
