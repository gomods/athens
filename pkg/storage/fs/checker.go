package fs

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/spf13/afero"
)

func (v *storageImpl) Exists(ctx context.Context, module, version string) (bool, error) {
	const op errors.Op = "fs.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	versionedPath := v.versionLocation(module, version)

	ok, err := afero.Exists(v.filesystem, versionedPath)
	if err != nil {
		return false, errors.E(op, errors.M(module), errors.V(version), err)
	}
	if !ok {
		return false, nil
	}

	files, err := afero.ReadDir(v.filesystem, versionedPath)
	if err != nil {
		return false, errors.E(op, errors.M(module), errors.V(version), err)
	}

	return len(files) == 3, nil
}
