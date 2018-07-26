package download

import (
	"net/http"

	"github.com/bketelsen/buffet"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gomods/athens/pkg/storage"
)

// PathVersionModule URL.
const PathVersionModule = "/{module:.+}/@v/{version}.mod"

// VersionModuleHandler implements GET baseURL/module/@v/version.mod
func VersionModuleHandler(dp download.Protocol, lggr *log.Logger, eng *render.Engine) buffalo.Handler {
	return func(c buffalo.Context, module, version string, versionInfo *storage.Version) error {
		sp := buffet.SpanFromContext(c)
		sp.SetOperationName("versionModuleHandler")
		c.Response().WriteHeader(http.StatusOK)
		_, err := c.Response().Write(versionInfo.Mod)
		return err
	}
}
