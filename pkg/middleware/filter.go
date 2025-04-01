package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gorilla/mux"
)

// NewFilterMiddleware builds a middleware function that implements the
// filters configured in the filter file.
func NewFilterMiddleware(mf *module.Filter, upstreamEndpoint string) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {
			mod, err := paths.GetModule(r)
			if err != nil {
				// if there is no module the path we are hitting is not one related to modules, like /
				h.ServeHTTP(w, r)
				return
			}
			ver, err := paths.GetVersion(r)
			if err != nil {
				ver = ""
			}
			rule := mf.Rule(mod, ver)
			switch rule {
			case module.Exclude:
				// Exclude: ignore request for this module
				w.WriteHeader(http.StatusForbidden)
				return
			case module.Include:
				// Include: please handle this module in a usual way
				h.ServeHTTP(w, r)
				return
			case module.Direct:
				// Direct: do not store modules locally, use upstream proxy
				newURL := redirectToUpstreamURL(upstreamEndpoint, r.URL)
				http.Redirect(w, r, newURL, http.StatusSeeOther)
				return
			}
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(f)
	}
}

func redirectToUpstreamURL(upstreamEndpoint string, u *url.URL) string {
	return strings.TrimSuffix(upstreamEndpoint, "/") + u.Path
}
