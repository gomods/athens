package actions

import (
	"github.com/gobuffalo/buffalo"
	"net/http"
)

func proxyHomeHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, proxy.JSON("Welcome to The Athens Proxy"))
}
