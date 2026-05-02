package download

import (
	"net/http"

	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
)

// PathVersionInfo URL.
const PathVersionInfo = "/{module:.+}/@v/{version}.info"

// InfoHandler implements GET baseURL/module/@v/version.info.
func InfoHandler(dp Protocol, lggr log.Entry, df *mode.DownloadFile) http.Handler {
	const op errors.Op = "download.InfoHandler"
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		mod, ver, err := getModuleParams(r, op)
		if err != nil {
			lggr.SystemErr(err)
			w.WriteHeader(errors.Kind(err))
			return
		}
		info, err := dp.Info(r.Context(), mod, ver)
		if err != nil {
			severityLevel := errors.Expect(err, errors.KindNotFound, errors.KindRedirect, errors.KindGone)
			lggr.SystemErr(errors.E(op, err, errors.M(mod), errors.V(ver), severityLevel))
			if errors.Kind(err) == errors.KindRedirect {
				url, err := getRedirectURL(df.URL(mod), r.URL.Path)
				if err != nil {
					lggr.SystemErr(err)
					w.WriteHeader(errors.Kind(err))
					return
				}
				http.Redirect(w, r, url, errors.KindRedirect)
				return
			}
			w.WriteHeader(errors.Kind(err))
			// For 410 Gone errors, include the descriptive error message in the response body
			// to help users understand what went wrong (e.g., semver mismatch)
			if errors.Kind(err) == errors.KindGone {
				_, _ = w.Write([]byte(err.Error()))
			}
		}

		_, _ = w.Write(info)
	}
	return http.HandlerFunc(f)
}
