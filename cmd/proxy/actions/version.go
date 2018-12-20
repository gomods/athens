package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/build"
	"net/http"
)

func versionHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, proxy.JSON(build.Data()))
}
