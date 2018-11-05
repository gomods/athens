package middleware

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/errors"
)

// NewUpstreamRedirectMiddleware redirects to upstream proxy if defined and result from local is not found
func NewUpstreamRedirectMiddleware(upstreamEndpoint string) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			result := next(c)
			if errors.Kind(result) == errors.KindNotFound && upstreamEndpoint != "" {
				newURL := redirectToRegistryURL(upstreamEndpoint, c.Request().URL)
				return c.Redirect(http.StatusSeeOther, newURL)
			}

			return result
		}
	}
}
