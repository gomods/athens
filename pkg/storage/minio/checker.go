package minio

import (
	"context"
	"fmt"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

const (
	minioErrorCodeNoSuchKey = "NoSuchKey"
)

func (v *storageImpl) Exists(ctx context.Context, module, version string) (bool, error) {
	const op errors.Op = "minio.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	versionedPath := v.versionLocation(module, version)
	modPath := fmt.Sprintf("%s/go.mod", versionedPath)
	infoPath := fmt.Sprintf("%s/%s.info", versionedPath, version)
	zipPath := fmt.Sprintf("%s/source.zip", versionedPath)

	var count int
	objectCh, _ := v.minioCore.ListObjectsV2(v.bucketName, versionedPath, "", false, "", 0, "")
	for _, object := range objectCh.Contents {
		if object.Err != nil {
			return false, errors.E(op, object.Err, errors.M(module), errors.V(version))
		}

		switch object.Key {
		case infoPath:
			count++
		case modPath:
			count++
		case zipPath:
			count++
		}
	}

	return count == 3, nil
}
