package minio

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	modupl "github.com/gomods/athens/pkg/storage/module"
	minio "github.com/minio/minio-go"
)

const (
	minioErrorCodeNoSuchKey = "NoSuchKey"
)

func (v *storageImpl) Exists(ctx context.Context, module, version string) (bool, error) {
	const op errors.Op = "minio.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	return modupl.Exists(ctx, module, version, func(ctx context.Context, name string) (bool, error) {
		_, err := v.minioClient.StatObject(v.bucketName, name, minio.StatObjectOptions{})

		if minio.ToErrorResponse(err).Code == minioErrorCodeNoSuchKey {
			return false, nil
		} else if err != nil {
			return false, errors.E(op, err, errors.M(module), errors.V(version))
		}
		return true, nil
	})
}
