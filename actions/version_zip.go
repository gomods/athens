package actions

import (
	"io"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/storage"
)

func versionZipHandler(getter storage.Getter) func(c buffalo.Context) error {
	return func(c buffalo.Context) error {
		params, err := getAllPathParams(c)
		if err != nil {
			return err
		}
		version, err := getter.Get(params.baseURL, params.module, params.version)
		if err != nil {
			return err
		}
		c.Response().WriteHeader(http.StatusOK)
		io.Copy(c.Response(), version.Zip)
		version.Zip.Close()
		return err
	}
}
