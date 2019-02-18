package minio

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	moduploader "github.com/gomods/athens/pkg/storage/module"
	minio "github.com/minio/minio-go"
)

func (s *storageImpl) Save(ctx context.Context, module, version string, mod []byte, zip storage.Zip, info []byte) error {
	const op errors.Op = "minio.Save"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	zipStream := moduploader.Stream{Stream: zip.Zip, Size: zip.Size}
	err := moduploader.Upload(ctx, module, version, moduploader.NewStreamFromBytes(info), moduploader.NewStreamFromBytes(mod), zipStream, s.upload, s.timeout)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	return nil
}

func (s *storageImpl) upload(ctx context.Context, path, contentType string, stream moduploader.Stream) error {
	const op errors.Op = "minio.upload"
	if stream.Size > 600000000 {
		stream.Size = -1
	}
	_, err := s.minioClient.PutObject(s.bucketName, path, stream.Stream, stream.Size, minio.PutObjectOptions{})
	return err
}
