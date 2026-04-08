package minio

import (
	"context"
	"fmt"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	minio "github.com/minio/minio-go/v7"
)

func (s *storageImpl) Delete(ctx context.Context, module, version string) error {
	const op errors.Op = "minio.Delete"

	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	exists, err := s.Exists(ctx, module, version)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	if !exists {
		return errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}

	versionedPath := s.versionLocation(module, version)

	modPath := fmt.Sprintf("%s/go.mod", versionedPath)

	err = s.minioClient.RemoveObject(ctx, s.bucketName, modPath, minio.RemoveObjectOptions{})
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	zipPath := fmt.Sprintf("%s/source.zip", versionedPath)

	err = s.minioClient.RemoveObject(ctx, s.bucketName, zipPath, minio.RemoveObjectOptions{})
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	infoPath := fmt.Sprintf("%s/%s.info", versionedPath, version)

	err = s.minioClient.RemoveObject(ctx, s.bucketName, infoPath, minio.RemoveObjectOptions{})
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	return nil
}
