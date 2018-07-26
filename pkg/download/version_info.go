package download

import (
	"encoding/json"
	"net/http"

	"github.com/bketelsen/buffet"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gomods/athens/pkg/storage"
)

// PathVersionInfo URL.
const PathVersionInfo = "/{module:.+}/@v/{version}.info"

// VersionInfoHandler implements GET baseURL/module/@v/version.info
func VersionInfoHandler(dp download.Protocol, lggr *log.Logger, eng *render.Engine) buffalo.Handler {
	return func(c buffalo.Context) error {
		sp := buffet.SpanFromContext(c)
		sp.SetOperationName("versionInfoHandler")
		var revInfo storage.RevInfo
		if err := json.Unmarshal(versionInfo.Info, &revInfo); err != nil {
			return err
		}
		return c.Render(http.StatusOK, eng.JSON(revInfo))
	}
}
