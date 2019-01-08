package minio

import (
	"bytes"
	"context"
	"io"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	moduploader "github.com/gomods/athens/pkg/storage/module"
	minio "github.com/minio/minio-go"
)

func (s *storageImpl) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	const op errors.Op = "minio.Save"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	err := moduploader.Upload(ctx, module, version, bytes.NewReader(info), bytes.NewReader(mod), zip, s.upload, 300)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	return nil
}

func (s *storageImpl) upload(ctx context.Context, path, contentType string, stream io.Reader) error {
	const op errors.Op = "minio.upload"
	_, err := s.minioClient.PutObject(s.bucketName, path, stream, -1, minio.PutObjectOptions{ContentType: contentType})
	return err
}
