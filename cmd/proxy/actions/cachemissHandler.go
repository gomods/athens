package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

func cachemissHandler(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		err := next(c)

		if httperr, ok := err.(buffalo.HTTPError); ok && httperr.Status == http.StatusNotFound {
			// TODO: report cache miss based on
		}

		return err
	}
}
