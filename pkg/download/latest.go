package download

import (
	"encoding/json"
	"net/http"

	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/paths"
)

// PathLatest URL.
const PathLatest = "/{module:.+}/@latest"

// LatestHandler implements GET baseURL/module/@latest
func LatestHandler(dp Protocol, lggr log.Entry, df *mode.DownloadFile) http.Handler {
	const op errors.Op = "download.LatestHandler"
	f := func(w http.ResponseWriter, r *http.Request) {
		mod, err := paths.GetModule(r)
		if err != nil {
			lggr.SystemErr(errors.E(op, err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		info, err := dp.Latest(r.Context(), mod)
		if err != nil {
			lggr.SystemErr(errors.E(op, err))
			w.WriteHeader(errors.Kind(err))
			return
		}

		if err = json.NewEncoder(w).Encode(info); err != nil {
			lggr.SystemErr(errors.E(op, err))
		}
	}
	return http.HandlerFunc(f)
}
