package middleware

import (
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/paths"
	"github.com/spf13/afero"
)

// NewPseudoversionMiddleware builds a middleware function that detects if the asked version
// is a hash and translates it into the mapped pseudoversion. It implements a workaround
// for https://github.com/golang/go/issues/27947
func NewPseudoversionMiddleware(fs afero.Fs, gobin string) buffalo.MiddlewareFunc {
	const op errors.Op = "actions.NewFilterMiddleware"

	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			mod, err := paths.GetModule(c)
			if err != nil {
				return next(c)
			}

			version, err := paths.GetVersion(c)
			if err != nil {
				return next(c)
			}

			if module.IsSemVersion(version) {
				return next(c)
			}

			newVersion, err := module.PseudoVersionFromHash(c, fs, gobin, mod, version)
			newURL := strings.Replace(c.Request().URL.Path, version, newVersion, 1)
			return c.Redirect(http.StatusSeeOther, newURL)
		}
	}
}
