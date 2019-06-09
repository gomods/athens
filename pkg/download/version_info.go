package download

import (
	"net/http"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
)

// PathVersionInfo URL.
const PathVersionInfo = "/{module:.+}/@v/{version}.info"

// InfoHandler implements GET baseURL/module/@v/version.info
func InfoHandler(dp Protocol, lggr log.Entry) http.Handler {
	const op errors.Op = "download.InfoHandler"
	f := func(w http.ResponseWriter, r *http.Request) {
		mod, ver, err := getModuleParams(r, op)
		if err != nil {
			lggr.SystemErr(err)
			w.WriteHeader(errors.Kind(err))
			return
		}
		info, err := dp.Info(r.Context(), mod, ver, lggr)
		if err != nil {
			lggr.SystemErr(errors.E(op, err, errors.M(mod), errors.V(ver)))
			w.WriteHeader(errors.Kind(err))
		}

		w.Write(info)
	}
	return http.HandlerFunc(f)
}
