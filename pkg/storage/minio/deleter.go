package minio

import (
	"context"
	"fmt"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

func (v *storageImpl) Delete(ctx context.Context, module, version string) error {
	const op errors.Op = "minio.Delete"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	exists, err := v.Exists(ctx, module, version)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	if !exists {
		return errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}

	versionedPath := v.versionLocation(module, version)

	modPath := fmt.Sprintf("%s.mod", versionedPath)
	if err := v.minioClient.RemoveObject(v.bucketName, modPath); err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	zipPath := fmt.Sprintf("%s.zip", versionedPath)
	if err := v.minioClient.RemoveObject(v.bucketName, zipPath); err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	infoPath := fmt.Sprintf("%s.info", versionedPath)
	err = v.minioClient.RemoveObject(v.bucketName, infoPath)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	return nil
}
