package actions

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gorilla/mux"
)

const defaultPageSize = 1000

type catalogRes struct {
	ModsAndVersions []paths.AllPathParams `json:"modules"`
	NextPageToken   string                `json:"next,omitempty"`
}

// catalogHandler implements GET baseURL/catalog
func catalogHandler(s storage.Backend) http.HandlerFunc {
	const op errors.Op = "actions.CatalogHandler"
	cs, isCataloger := s.(storage.Cataloger)
	f := func(w http.ResponseWriter, r *http.Request) {
		if !isCataloger {
			w.WriteHeader(errors.KindNotImplemented)
			return
		}
		lggr := log.EntryFromContext(r.Context())
		vars := mux.Vars(r)
		token := vars["token"]
		pageSize, err := getLimitFromParam(vars["pagesize"])
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

func getLimitFromParam(param string) (int, error) {
	if param == "" {
		return defaultPageSize, nil
	}
	return strconv.Atoi(param)
}
