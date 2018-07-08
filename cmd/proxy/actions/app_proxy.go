package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/config/env"
	"github.com/gomods/athens/pkg/storage"
)

func addProxyRoutes(app *buffalo.App, storage storage.Backend) error {
	app.GET("/", proxyHomeHandler)
	app.GET("/{module:.+}/@v/list", listHandler(storage))
	app.GET("/{module:.+}/@v/{version}.info", cacheMissHandler(versionInfoHandler(storage), app.Worker))
	app.GET("/{module:.+}/@v/{version}.mod", cacheMissHandler(versionModuleHandler(storage), app.Worker))
	app.GET("/{module:.+}/@v/{version}.zip", cacheMissHandler(versionZipHandler(storage), app.Worker))

	// Only allow proxy module uploads during development.
	if env.GoEnvironmentWithDefault("") == "development" {
		app.POST("/admin/upload/{module:.+}/{version}", uploadHandler(storage))
	}

	app.POST("/admin/fetch/{module:.+}/{owner}/{repo}/{ref}/{version}", fetchHandler(storage))
	return nil
}
