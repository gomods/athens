package fs

import (
	"context"
	"os"
	"path/filepath"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
)

func (v *storageImpl) Info(ctx context.Context, module, version string) ([]byte, error) {
	const op errors.Op = "fs.Info"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	versionedPath := v.versionLocation(module, version)
	info, err := afero.ReadFile(v.filesystem, filepath.Join(versionedPath, version+".info"))
	if err != nil {
		return nil, errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}

	return info, nil
}

func (v *storageImpl) GoMod(ctx context.Context, module, version string) ([]byte, error) {
	const op errors.Op = "fs.GoMod"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	versionedPath := v.versionLocation(module, version)
	mod, err := afero.ReadFile(v.filesystem, filepath.Join(versionedPath, "go.mod"))
	if err != nil {
		return nil, errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}

	return mod, nil
}

func (v *storageImpl) Zip(ctx context.Context, module, version string) (storage.SizeReadCloser, error) {
	const op errors.Op = "fs.Zip"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	versionedPath := v.versionLocation(module, version)

	src, err := v.filesystem.OpenFile(filepath.Join(versionedPath, "source.zip"), os.O_RDONLY, 0666)
	if err != nil {
		return nil, errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}
	fi, err := src.Stat()
	if err != nil {
		return nil, errors.E(op, err)
	}
	return storage.NewSizer(src, fi.Size()), nil
}

func (v *storageImpl) ZipSize(ctx context.Context, module, version string) (int64, error) {
	const op errors.Op = "fs.ZipFileSize"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	versionedPath := v.versionLocation(module, version)
	fi, err := v.filesystem.Stat(filepath.Join(versionedPath))
	if err != nil {
		return 0, errors.E(op, err, errors.M(module), errors.V(version), errors.KindNotFound)
	}
	return fi.Size(), nil
}
