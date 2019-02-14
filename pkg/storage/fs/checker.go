package fs

import (
	"context"
	"fmt"
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
	files := []string{
		"go.mod",
		fmt.Sprintf("%s.info", version),
		fmt.Sprintf("%s.zip", version),
	}
	for _, file := range files {
		exists, err := afero.Exists(v.filesystem, filepath.Join(versionedPath, file))
		if err != nil {
			return false, errors.E(op, errors.M(module), errors.V(version), err)
		}
		if !exists {
			return false, nil
		}
	}

	return true, nil
}
