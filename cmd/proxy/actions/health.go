package actions

import (
	"github.com/gobuffalo/buffalo"
	"net/http"
)

func healthHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, nil)
}
