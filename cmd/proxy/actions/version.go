package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/build"
)

func versionHandler(c buffalo.Context) error {
	return c.Render(200, proxy.JSON(build.Data()))
}
