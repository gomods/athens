package download

import (
	"net/http"
	"strconv"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/paths"
)

// PathCatalog URL.
const PathCatalog = "/catalog"
const defaultPageSize = 1000

type catalogRes struct {
	ModsAndVersions []paths.AllPathParams `json:"modules"`
	NextPageToken   string                `json:"next,omitempty"`
}

// CatalogHandler implements GET baseURL/catalog
func CatalogHandler(dp Protocol, lggr log.Entry, eng *render.Engine) buffalo.Handler {
	const op errors.Op = "download.CatalogHandler"

	return func(c buffalo.Context) error {
		token := c.Param("token")
		pageSize, err := getLimitFromParam(c.Param("pagesize"))
		if err != nil {
			lggr.SystemErr(err)
			return c.Render(http.StatusInternalServerError, nil)
		}

		modulesAndVersions, newToken, err := dp.Catalog(c, token, pageSize)

		if err != nil {
			if errors.Kind(err) != errors.KindNotImplemented {
				lggr.SystemErr(errors.E(op, err))
			}
			return c.Render(errors.Kind(err), eng.JSON(errors.KindText(err)))
		}

		res := catalogRes{modulesAndVersions, newToken}
		return c.Render(http.StatusOK, eng.JSON(res))
	}
}

func getLimitFromParam(param string) (int, error) {
	if param == "" {
		return defaultPageSize, nil
	}
	return strconv.Atoi(param)
}
