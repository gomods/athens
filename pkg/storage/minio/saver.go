package minio

import (
	"context"
	"io"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	moduploader "github.com/gomods/athens/pkg/storage/module"
	minio "github.com/minio/minio-go"
)

func (s *storageImpl) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte, size int64) error {
	const op errors.Op = "minio.Save"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	err := moduploader.Upload(ctx, module, version, moduploader.NewStreamFromBytes(info), moduploader.NewStreamFromBytes(mod), moduploader.NewStreamFromReaderWithSize(zip, size), s.upload, s.timeout)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	return nil
}

func (s *storageImpl) upload(ctx context.Context, path, contentType string, stream moduploader.Stream) error {
	const op errors.Op = "minio.upload"
	_, err := s.minioClient.PutObject(s.bucketName, path, stream.Stream, stream.Size, minio.PutObjectOptions{})
	return err
}
