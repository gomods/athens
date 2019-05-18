package download

import (
	"net/http"

	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
)

// PathVersionInfo URL.
const PathVersionInfo = "/{module:.+}/@v/{version}.info"

// InfoHandler implements GET baseURL/module/@v/version.info
func InfoHandler(dp Protocol, lggr log.Entry, df *mode.DownloadFile) http.Handler {
	const op errors.Op = "download.InfoHandler"
	f := func(w http.ResponseWriter, r *http.Request) {
		mod, ver, err := getModuleParams(r, op)
		if err != nil {
			lggr.SystemErr(err)
			w.WriteHeader(errors.Kind(err))
			return
		}
		info, err := dp.Info(r.Context(), mod, ver)
		if err != nil {
			lggr.SystemErr(errors.E(op, err, errors.M(mod), errors.V(ver)))
			if errors.Kind(err) == errors.KindRedirect {
				http.Redirect(w, r, getRedirectURL(df.URL(mod), r.URL.Path), errors.KindRedirect)
				return
			}
			w.WriteHeader(errors.Kind(err))
		}

		w.Write(info)
	}
	return http.HandlerFunc(f)
}
