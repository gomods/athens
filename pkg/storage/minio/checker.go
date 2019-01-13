package minio

import (
	"context"
	"fmt"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	minio "github.com/minio/minio-go"
)

const (
	minioErrorCodeNoSuchKey = "NoSuchKey"
)

func (v *storageImpl) Exists(ctx context.Context, module, version string) (bool, error) {
	const op errors.Op = "minio.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	versionedPath := v.versionLocation(module, version)
	modPath := fmt.Sprintf("%s.mod", versionedPath)
	_, err := v.minioClient.StatObject(v.bucketName, modPath, minio.StatObjectOptions{})

	if minio.ToErrorResponse(err).Code == minioErrorCodeNoSuchKey {
		return false, nil
	}

	if err != nil {
		return false, errors.E(op, err, errors.M(module), errors.V(version))
	}

	return true, nil
}
