package download

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/paths"
)

// PathLatest URL.
const PathLatest = "/{module:.+}/@latest"

// LatestHandler implements GET baseURL/module/@latest
func LatestHandler(dp Protocol, eng *render.Engine) buffalo.Handler {
	const op errors.Op = "download.LatestHandler"
	return func(c buffalo.Context) error {
		mod, err := paths.GetModule(c)
		if err != nil {
			c.Logger().Warn(errors.E(op, err))
			// lggr.SystemErr(errors.E(op, err))
			return c.Render(500, nil)
		}

		info, err := dp.Latest(c, mod)
		if err != nil {
			c.Logger().Warn(errors.E(op, err))
			// lggr.SystemErr(errors.E(op, err))
			return c.Render(errors.Kind(err), eng.JSON(errors.KindText(err)))
		}

		return c.Render(http.StatusOK, eng.JSON(info))
	}
}
