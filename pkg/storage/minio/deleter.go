package minio

import (
	"fmt"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

func (v *storageImpl) Delete(ctx observ.ProxyContext, module, version string) error {
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

	modPath := fmt.Sprintf("%s/go.mod", versionedPath)
	if err := v.minioClient.RemoveObject(v.bucketName, modPath); err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	zipPath := fmt.Sprintf("%s/source.zip", versionedPath)
	if err := v.minioClient.RemoveObject(v.bucketName, zipPath); err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	infoPath := fmt.Sprintf("%s/%s.info", versionedPath, version)
	err = v.minioClient.RemoveObject(v.bucketName, infoPath)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	return nil
}
