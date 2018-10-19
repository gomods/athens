package contenttype

import "github.com/gobuffalo/buffalo"

// Set on the request to desired type. This will
// override any content type sent by the client.
func Set(s string) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			c.Request().Header.Set("Content-Type", s)
			return next(c)
		}
	}
}

// Add will add a secondary content type to
// a request. If no content type is sent by the client
// the default will be set, otherwise the client's
// content type will be used.
func Add(s string) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			c.Request().Header.Add("Content-Type", s)
			return next(c)
		}
	}
}
