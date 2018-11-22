package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/storage"
	"net"
)

func getReadinessHandler(s storage.Backend) buffalo.Handler {
	return func(c buffalo.Context) error {
		_, err := s.List(c, "github.com/gomods/athens")

		if err != nil {
			return c.Render(500, nil)
		}

		_, err = net.LookupIP("github.com")
		if err != nil {
			return c.Render(500, nil)
		}

		return c.Render(200, nil)
	}
}
