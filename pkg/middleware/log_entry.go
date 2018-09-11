package middleware

import (
	"github.com/gobuffalo/buffalo"
	"github.com/sirupsen/logrus"
)

// LogEntryMiddleware builds a log.Entry applying the request parameter to the given
// log.Logger and propagates it to the given MiddlewareFunc
func LogEntryMiddleware(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		req := c.Request()
		c.LogFields(logrus.Fields{
			"http-method": req.Method,
			"http-path":   req.URL.Path,
			"http-url":    req.URL.String(),
		})
		return next(c)
	}
}
