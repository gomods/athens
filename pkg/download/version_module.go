package download

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// PathVersionModule URL.
const PathVersionModule = "/{module:.+}/@v/{version}.mod"

// VersionModuleHandler implements GET baseURL/module/@v/version.mod
func VersionModuleHandler(dp Protocol, eng *render.Engine) buffalo.Handler {
	const op errors.Op = "download.VersionModuleHandler"
	return func(c buffalo.Context) error {
		ctx, span := observ.StartSpan(c, op.String())
		defer span.End()
		mod, ver, err := getModuleParams(c, op)
		if err != nil {
			c.Logger().Warn(errors.E(op, err))
			// lggr.SystemErr(err)
			return c.Render(errors.Kind(err), nil)
		}
		modBts, err := dp.GoMod(ctx, mod, ver)
		if err != nil {
			err = errors.E(op, errors.M(mod), errors.V(ver), err)
			c.Logger().Warn(err)
			// lggr.SystemErr(err)
			return c.Render(errors.Kind(err), nil)
		}

		// Calling c.Response().Write will write the header directly
		// and we would get a 0 status in the buffalo logs.
		return c.Render(200, eng.String(string(modBts)))
	}
}
