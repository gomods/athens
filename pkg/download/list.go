package download

import (
	"context"
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/paths"
)

// PathList URL.
const PathList = "/{module:.+}/@v/list"

// ListHandler implements GET baseURL/module/@v/list
func ListHandler(dp Protocol, lggr log.Entry, eng *render.Engine) buffalo.Handler {
	const op errors.Op = "download.ListHandler"
	return func(c buffalo.Context) error {
		mod, err := paths.GetModule(c)
		if err != nil {
			lggr.SystemErr(errors.E(op, err))
			return c.Render(500, nil)
		}

		versions, err := dp.List(c, mod)
		if err != nil {
			lggr.SystemErr(errors.E(op, err))
			return c.Render(errors.Kind(err), eng.JSON(errors.KindText(err)))
		}

		return c.Render(http.StatusOK, eng.String(strings.Join(versions, "\n")))
	}
}

// ListHandlerBasic implements GET baseURL/module/@v/list as a basic http.HandlerFunc
// wrapping it as a buffalo handler
func ListHandlerBasic(dp Protocol, lggr log.Entry, eng *render.Engine) http.HandlerFunc {
	const op errors.Op = "download.ListHandler"
	return func(w http.ResponseWriter, r *http.Request) {

		mod, err := paths.GetModuleFromRequest(r)
		if err != nil {
			lggr.SystemErr(errors.E(op, err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		versions, err := dp.List(context.Background(), mod)
		if err != nil {
			lggr.SystemErr(errors.E(op, err))
			http.Error(w, errors.KindText(err), errors.Kind(err))
			return
		}

		w.Write([]byte(strings.Join(versions, "\n")))
	}
}
