package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/storage"
	"net/http"
)

func getReadinessHandler(s storage.Backend) buffalo.Handler {
	return func(c buffalo.Context) error {
		if _, err := s.List(c, "github.com/gomods/athens"); err != nil {
			return c.Render(http.StatusInternalServerError, nil)
		}

		return c.Render(http.StatusOK, nil)
	}
}
