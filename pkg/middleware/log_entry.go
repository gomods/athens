package middleware

import (
	"net/http"

	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/requestid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// LogEntryMiddleware builds a log.Entry, setting the request fields
// and storing it in the context to be used throughout the stack
func LogEntryMiddleware(lggr *log.Logger) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ent := lggr.WithFields(logrus.Fields{
				"http-method": r.Method,
				"http-path":   r.URL.Path,
				"request-id":  requestid.FromContext(ctx),
			})
			ctx = log.SetEntryInContext(ctx, ent)
			r = r.WithContext(ctx)
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(f)
	}
}
