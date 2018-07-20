package download

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bketelsen/buffet"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gomods/athens/pkg/storage"
)

// PathVersionInfo URL.
const PathVersionInfo = "/{module:.+}/@v/{version}.info"

// VersionInfoHandler implements GET baseURL/module/@v/version.info
func VersionInfoHandler(
	getterSaver storage.GetterSaver,
	fetcher module.Fetcher,
	eng *render.Engine,
) func(c buffalo.Context) error {
	return func(c buffalo.Context) error {
		sp := buffet.SpanFromContext(c)
		sp.SetOperationName("versionInfoHandler")
		params, err := paths.GetAllParams(c)
		if err != nil {
			return err
		}
		version, err := getterSaver.Get(params.Module, params.Version)
		if storage.IsNotFoundError(err) {
			// TODO: serialize cache fills (https://github.com/gomods/athens/issues/308)
			ref, err := fetcher.Fetch(params.Module, params.Version)
			if err != nil {
				// TODO: some way to figure out whether the package actually doesn't exist
				return err
			}
			defer ref.Clear()
			ver, err := ref.Read()
			if err != nil {
				return err
			}
			getterSaver.Save(c, params.Module, params.Version, ver.Mod, ver.Zip, ver.Info)
			version = ver
		} else if err != nil {
			return err
		}
		revInfo, err := parseRevInfo(version.Info)
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, eng.JSON(revInfo))
	}
}

func parseRevInfo(b []byte) (*storage.RevInfo, error) {
	revInfo := &storage.RevInfo{}
	if err := json.Unmarshal(b, revInfo); err != nil {
		return nil, err
	}
	return revInfo, nil
}
