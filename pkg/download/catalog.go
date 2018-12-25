package download

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gorilla/mux"
)

// PathCatalog URL.
const PathCatalog = "/catalog"
const defaultPageSize = 1000

type catalogRes struct {
	ModsAndVersions []paths.AllPathParams `json:"modules"`
	NextPageToken   string                `json:"next"`
}

// CatalogHandler implements GET baseURL/catalog
func CatalogHandler(dp Protocol, lggr log.Entry) http.Handler {
	const op errors.Op = "download.CatalogHandler"
	f := func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		token := vars["token"]
		pageSize, err := getLimitFromParam(vars["pagesize"])
		if err != nil {
			lggr.SystemErr(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		modulesAndVersions, newToken, err := dp.Catalog(r.Context(), token, pageSize)

		if err != nil {
			if errors.Kind(err) != errors.KindNotImplemented {
				lggr.SystemErr(errors.E(op, err))
			}
			w.WriteHeader(errors.Kind(err))
			return
		}

		res := catalogRes{modulesAndVersions, newToken}
		if err = json.NewEncoder(w).Encode(res); err != nil {
			lggr.SystemErr(errors.E(op, err))
		}
	}
	return http.HandlerFunc(f)
}

func getLimitFromParam(param string) (int, error) {
	if param == "" {
		return defaultPageSize, nil
	}
	return strconv.Atoi(param)
}
