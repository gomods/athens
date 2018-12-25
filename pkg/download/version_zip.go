package download

import (
	"io"
	"net/http"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
)

// PathVersionZip URL.
const PathVersionZip = "/{module:.+}/@v/{version}.zip"

// ZipHandler implements GET baseURL/module/@v/version.zip
func ZipHandler(dp Protocol, lggr log.Entry) http.Handler {
	const op errors.Op = "download.ZipHandler"
	f := func(w http.ResponseWriter, r *http.Request) {
		mod, ver, err := getModuleParams(r, op)
		if err != nil {
			lggr.SystemErr(err)
			w.WriteHeader(errors.Kind(err))
			return
		}
		zip, err := dp.Zip(r.Context(), mod, ver)
		if err != nil {
			lggr.SystemErr(err)
			w.WriteHeader(errors.Kind(err))
			return
		}
		defer zip.Close()

		_, err = io.Copy(w, zip)
		if err != nil {
			lggr.SystemErr(errors.E(op, errors.M(mod), errors.V(ver), err))
		}
	}
	return http.HandlerFunc(f)
}
