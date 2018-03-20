package actions

import (
	"github.com/gobuffalo/buffalo"
)

func proxyHomeHandler(c buffalo.Context) error {
	c.Flash().Add("info", "Proxy")
	r.HTMLLayout = "proxy/application.html"
	return c.Render(200, r.HTML("proxy/index.html"))
}

func homeHandler(c buffalo.Context) error {
	c.Flash().Add("info", "Registry")
	r.HTMLLayout = "registry/application.html"
	return c.Render(200, r.HTML("registry/index.html"))
}
