package minio

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	minio "github.com/minio/minio-go/v6"
)

func (s *storageImpl) Info(ctx context.Context, module, vsn string) ([]byte, error) {
	const op errors.Op = "minio.Info"
	_, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	infoPath := fmt.Sprintf("%s/%s.info", s.versionLocation(module, vsn), vsn)
	infoReader, err := s.minioClient.GetObject(s.bucketName, infoPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer func() { _ = infoReader.Close() }()
	info, err := io.ReadAll(infoReader)
	if err != nil {
		return nil, transformNotFoundErr(op, module, vsn, err)
	}

	return info, nil
}

func (s *storageImpl) GoMod(ctx context.Context, module, vsn string) ([]byte, error) {
	const op errors.Op = "minio.GoMod"
	_, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	modPath := fmt.Sprintf("%s/go.mod", s.versionLocation(module, vsn))
	modReader, err := s.minioClient.GetObject(s.bucketName, modPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer func() { _ = modReader.Close() }()
	mod, err := io.ReadAll(modReader)
	if err != nil {
		return nil, transformNotFoundErr(op, module, vsn, err)
	}

	return mod, nil
}

func (s *storageImpl) Zip(ctx context.Context, module, vsn string) (storage.SizeReadCloser, error) {
	const op errors.Op = "minio.Zip"
	_, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	zipPath := fmt.Sprintf("%s/source.zip", s.versionLocation(module, vsn))
	_, err := s.minioClient.StatObject(s.bucketName, zipPath, minio.StatObjectOptions{})
	if err != nil {
		return nil, errors.E(op, err, errors.KindNotFound, errors.M(module), errors.V(vsn))
	}

	zipReader, err := s.minioClient.GetObject(s.bucketName, zipPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.E(op, err)
	}
	oi, err := zipReader.Stat()
	if err != nil {
		_ = zipReader.Close()
		return nil, errors.E(op, err)
	}
	return storage.NewSizer(zipReader, oi.Size), nil
}

func transformNotFoundErr(op errors.Op, module, version string, err error) error {
	var eresp minio.ErrorResponse
	if errors.AsErr(err, &eresp) {
		if eresp.StatusCode == http.StatusNotFound {
			return errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
		}
	}
	return err
}
