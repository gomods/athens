package download

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gomods/athens/pkg/storage"
)

// Middleware ensures that next is called with the module@version in the cache.
func Middleware(
	next Handler,
	fetcher module.Fetcher,
	getterSaver storage.GetterSaver,
	lggr *log.Logger,
) buffalo.Handler {
	const op errors.Op = "download.fetchMiddleware"
	return func(c buffalo.Context) error {
		params, err := paths.GetAllParams(c)
		if err != nil {
			return errors.E(op, err)
		}
		module, version := params.Module, params.Version
		versionInfo, err := getterSaver.Get(module, version)
		if storage.IsNotFoundError(err) {
			ref, err := fetcher.Fetch(module, version)
			if err != nil {
				err := errors.E(op, errors.M(module), errors.V(version), err)
				lggr.SystemErr(err)
				return err
			}
			defer ref.Clear()
			versionInfo, err = ref.Read()
			if err != nil {
				err := errors.E(op, errors.M(module), errors.V(version), err)
				lggr.SystemErr(err)
				return err
			}
			defer versionInfo.Zip.Close()
			if err := getterSaver.Save(
				c,
				module,
				version,
				versionInfo.Mod,
				versionInfo.Zip,
				versionInfo.Info,
			); err != nil {
				err := errors.E(op, errors.M(module), errors.V(version), err)
				lggr.SystemErr(err)
				return err
			}
		} else if err != nil {
			err := errors.E(op, err)
			lggr.SystemErr(err)
			return err
		}
		return next(c, module, version, versionInfo)
	}
}
