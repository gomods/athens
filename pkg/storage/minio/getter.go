package minio

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	minio "github.com/minio/minio-go/v6"
)

func (v *storageImpl) Info(ctx context.Context, module, vsn string) ([]byte, error) {
	const op errors.Op = "minio.Info"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	infoPath := fmt.Sprintf("%s/%s.info", v.versionLocation(module, vsn), vsn)
	infoReader, err := v.minioClient.GetObject(v.bucketName, infoPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.E(op, err)
	}
	info, err := ioutil.ReadAll(infoReader)
	if err != nil {
		return nil, transformNotFoundErr(op, module, vsn, err)
	}

	return info, nil
}

func (v *storageImpl) GoMod(ctx context.Context, module, vsn string) ([]byte, error) {
	const op errors.Op = "minio.GoMod"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	modPath := fmt.Sprintf("%s/go.mod", v.versionLocation(module, vsn))
	modReader, err := v.minioClient.GetObject(v.bucketName, modPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.E(op, err)
	}
	mod, err := ioutil.ReadAll(modReader)
	if err != nil {
		return nil, transformNotFoundErr(op, module, vsn, err)
	}

	return mod, nil
}
func (v *storageImpl) Zip(ctx context.Context, module, vsn string) (storage.SizeReadCloser, error) {
	const op errors.Op = "minio.Zip"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	zipPath := fmt.Sprintf("%s/source.zip", v.versionLocation(module, vsn))
	_, err := v.minioClient.StatObject(v.bucketName, zipPath, minio.StatObjectOptions{})
	if err != nil {
		return nil, errors.E(op, err, errors.KindNotFound, errors.M(module), errors.V(vsn))
	}

	zipReader, err := v.minioClient.GetObject(v.bucketName, zipPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.E(op, err)
	}
	oi, err := zipReader.Stat()
	if err != nil {
		return nil, errors.E(op, err)
	}
	return storage.NewSizer(zipReader, oi.Size), nil
}

func transformNotFoundErr(op errors.Op, module, version string, err error) error {
	if eresp, ok := err.(minio.ErrorResponse); ok {
		if eresp.StatusCode == http.StatusNotFound {
			return errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
		}
	}
	return err
}
