package actions

import (
	"strings"

	"github.com/bketelsen/buffet"
	"github.com/fedepaol/athens/pkg/errors"
	"github.com/fedepaol/athens/pkg/log"
	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/paths"
)

func newFilterMiddleware(mf *module.Filter, lggr *log.Logger) buffalo.MiddlewareFunc {
	const op errors.Op = "actions.FilterMiddleware"

	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			sp := buffet.SpanFromContext(c).SetOperationName("filterMiddleware")
			defer sp.Finish()

			params, err := paths.GetAllParams(c)
			if err != nil {
				lggr.SystemErr(errors.E(op, err))
				return err
			}

			if isPseudoVersion(params.Version) {
				return next(c)
			}

			rule := mf.Rule(params.Module)
			switch rule {
			case module.Exclude:
				return module.NewErrModuleExcluded(params.Module)
			case module.Private:
				return next(c)
			case module.Include:
				return c.Redirect(303, GetOlympusEndpoint())
			}

			return next(c)
		}
	}
}

func isPseudoVersion(version string) bool {
	return strings.HasPrefix(version, "v0.0.0-")
}
