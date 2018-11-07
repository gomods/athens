package middleware

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/log"
	"github.com/sirupsen/logrus"
)

// LogEntryMiddleware builds a log.Entry, setting the request fields
// and storing it in the context to be used throughout the stack
func LogEntryMiddleware(lggr *log.Logger) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			req := c.Request()
			ent := lggr.WithFields(logrus.Fields{
				"http-method": req.Method,
				"http-path":   req.URL.Path,
				"http-url":    req.URL.String(),
			})

			log.SetEntryInContext(c, ent)
			return next(c)
		}
	}
}
