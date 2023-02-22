package fs

import (
	"context"
	"os"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/spf13/afero"
)

func (s *storageImpl) Exists(ctx context.Context, module, version string) (bool, error) {
	const op errors.Op = "fs.Exists"
	_, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	versionedPath := s.versionLocation(module, version)

	files, err := afero.ReadDir(s.filesystem, versionedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.E(op, errors.M(module), errors.V(version), err)
	}

	return len(files) == 3, nil
}
