package minio

import (
	"context"
	"fmt"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

func (s *storageImpl) Exists(ctx context.Context, module, version string) (bool, error) {
	const op errors.Op = "minio.Exists"
	_, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	versionedPath := s.versionLocation(module, version)
	modPath := fmt.Sprintf("%s/go.mod", versionedPath)
	infoPath := fmt.Sprintf("%s/%s.info", versionedPath, version)
	zipPath := fmt.Sprintf("%s/source.zip", versionedPath)

	var count int
	objectCh, err := s.minioCore.ListObjectsV2(s.bucketName, versionedPath, "", false, "", 0, "")
	if err != nil {
		return false, errors.E(op, err, errors.M(module), errors.V(version))
	}
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
