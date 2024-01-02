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

func (s *storageImpl) Info(ctx context.Context, module, version string) ([]byte, error) {
	const op errors.Op = "fs.Info"
	_, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	versionedPath := s.versionLocation(module, version)
	info, err := afero.ReadFile(s.filesystem, filepath.Join(versionedPath, version+".info"))
	if err != nil {
		return nil, errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}

	return info, nil
}

func (s *storageImpl) GoMod(ctx context.Context, module, version string) ([]byte, error) {
	const op errors.Op = "fs.GoMod"
	_, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	versionedPath := s.versionLocation(module, version)
	mod, err := afero.ReadFile(s.filesystem, filepath.Join(versionedPath, "go.mod"))
	if err != nil {
		return nil, errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}

	return mod, nil
}

func (s *storageImpl) Zip(ctx context.Context, module, version string) (storage.SizeReadCloser, error) {
	const op errors.Op = "fs.Zip"
	_, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	versionedPath := s.versionLocation(module, version)

	src, err := s.filesystem.OpenFile(filepath.Join(versionedPath, "source.zip"), os.O_RDONLY, 0o666)
	if err != nil {
		return nil, errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}
	fi, err := src.Stat()
	if err != nil {
		return nil, errors.E(op, err)
	}
	return storage.NewSizer(src, fi.Size()), nil
}
