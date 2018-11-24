package download

import (
	"net/http"
	"strconv"

	"github.com/gomods/athens/pkg/paths"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
)

// PathCatalog URL.
const PathCatalog = "/catalog"

type catalogRes struct {
	ModsAndVersions []paths.AllPathParams
	NextPageToken   string
}

// CatalogHandler implements GET baseURL/catalog
func CatalogHandler(dp Protocol, lggr log.Entry, eng *render.Engine) buffalo.Handler {
	const op errors.Op = "download.CatalogHandler"

	return func(c buffalo.Context) error {
		token := c.Param("token")
		limit, err := getLimitFromParam(c.Param("limit"))
		if err != nil {
			lggr.SystemErr(err)
			return c.Render(http.StatusInternalServerError, nil)
		}

		modulesAndVersions, newToken, err := dp.Catalog(c, token, limit)
		if err != nil {
			lggr.SystemErr(err)
			return c.Render(errors.Kind(err), nil)
		}
		if err != nil {
			lggr.SystemErr(errors.E(op, err))
			return c.Render(errors.Kind(err), eng.JSON(errors.KindText(err)))
		}

		res := catalogRes{modulesAndVersions, newToken}
		return c.Render(http.StatusOK, eng.JSON(res))
	}
}

func getLimitFromParam(param string) (int, error) {
	if param == "" {
		return 0, nil
	}
	return strconv.Atoi(param)

}
