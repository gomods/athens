package download

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/paths"
)

// PathList URL.
const PathList = "/{module:.+}/@v/list"

// ListHandler implements GET baseURL/module/@v/list
func ListHandler(dp Protocol, lggr log.Entry, df *mode.DownloadFile) http.Handler {
	const op errors.Op = "download.ListHandler"
	f := func(w http.ResponseWriter, r *http.Request) {
		mod, err := paths.GetModule(r)
		if err != nil {
			lggr.SystemErr(errors.E(op, err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		versions, err := dp.List(r.Context(), mod)
		if err != nil {
			lggr.SystemErr(errors.E(op, err))
			w.WriteHeader(errors.Kind(err))
			return
		}

		fmt.Fprint(w, strings.Join(versions, "\n"))
	}
	return http.HandlerFunc(f)
}
