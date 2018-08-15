package module

import (
	"context"

	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/errors"
)

// Fetch downloads the module@version using the fetcher and stores it in storage
func Fetch(ctx context.Context, s storage.Backend, fetcher Fetcher, mod, version string, mf *Filter) error {
	const op errors.Op = "module.Fetch"
	if !mf.ShouldProcess(mod) {
		return NewErrModuleExcluded(mod)
	}

	moduleExists, err := s.Exists(ctx, mod, version)
	if err != nil {
		return errors.E(op, err)
	}
	if moduleExists {
		return NewErrModuleAlreadyFetched("module.Fetch", mod, version)
	}

	moduleRef, err := fetcher.Fetch(mod, version)
	if err != nil {
		return errors.E(op, err)
	}

	// pretend like moduleLoc has $version.info, $version.mod and $version.zip in it :)
	module, err := moduleRef.Read()
	if err != nil {
		return errors.E(op, err)
	}
	defer module.Zip.Close()

	return s.Save(context.Background(), mod, version, module.Mod, module.Zip, module.Info)
}
