package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/storage"
)

func getReadinessHandler(s storage.Backend) buffalo.Handler {
	return func(c buffalo.Context) error {
		if _, err := s.List(c, "github.com/gomods/athens"); err != nil {
			return c.Render(500, nil)
		}

		return c.Render(200, nil)
	}
}
