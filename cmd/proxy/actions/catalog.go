package actions

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gomods/athens/pkg/storage"
)

const defaultPageSize = 1000

type catalogRes struct {
	ModsAndVersions []paths.AllPathParams `json:"modules"`
	NextPageToken   string                `json:"next,omitempty"`
}

// catalogHandler implements GET baseURL/catalog.
func catalogHandler(s storage.Backend) http.HandlerFunc {
	const op errors.Op = "actions.CatalogHandler"
	cs, isCataloger := s.(storage.Cataloger)
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if !isCataloger {
			w.WriteHeader(errors.KindNotImplemented)
			return
		}

		lggr := log.EntryFromContext(r.Context())
		token := r.FormValue("token")

		pageSize, err := getLimitFromParam(r.FormValue("pagesize"))
		if err != nil {
			lggr.SystemErr(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		modulesAndVersions, newToken, err := cs.Catalog(r.Context(), token, pageSize)
		if err != nil {
			lggr.SystemErr(errors.E(op, err))
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

// getLimitFromParam converts a URL query parameter into an int
// otherwise converts defaultPageSize constant.
func getLimitFromParam(param string) (int, error) {
	if param == "" {
		return defaultPageSize, nil
	}
	return strconv.Atoi(param)
}
