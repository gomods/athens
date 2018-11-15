package fs

import (
	"context"
	"path/filepath"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/spf13/afero"
)

func (v *storageImpl) Exists(ctx context.Context, module, version string) (bool, error) {
	const op errors.Op = "fs.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	versionedPath := v.versionLocation(module, version)
	exists, err := afero.Exists(v.filesystem, filepath.Join(versionedPath, "go.mod"))
	if err != nil {
		return false, errors.E(op, errors.M(module), errors.V(version), err)
	}

	return exists, nil
}
