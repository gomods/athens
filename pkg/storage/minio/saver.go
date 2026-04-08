package minio

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/hashicorp/go-multierror"
	minio "github.com/minio/minio-go/v7"
)

func (s *storageImpl) Save(ctx context.Context, module, vsn string, mod []byte, zip io.Reader, zipMD5, info []byte) error {
	const op errors.Op = "storage.minio.Save"

	_, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	dir := s.versionLocation(module, vsn)
	modFileName := dir + "/" + "go.mod"
	infoFileName := dir + "/" + vsn + ".info"

	_, err := s.minioClient.PutObject(ctx, s.bucketName, modFileName, bytes.NewReader(mod), int64(len(mod)), minio.PutObjectOptions{})
	if err != nil {
		return errors.E(op, err)
	}
	// Chunk the stream into 8mb and send them in parts to minio.
	// This is because the minio client over-allocates a stream buffer (600Mb)
	// when the size is unknown, see https://github.com/minio/minio-go/issues/848
	err = s.saveZip(ctx, dir, module, vsn, zip)
	if err != nil {
		return errors.E(op, err)
	}

	_, err = s.minioClient.PutObject(ctx, s.bucketName, infoFileName, bytes.NewReader(info), int64(len(info)), minio.PutObjectOptions{})
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

type partWriter struct {
	numParts   int
	c          *minio.Client
	mod, ver   string
	bucketName string
	err        error
	srcs       []minio.CopySrcOptions
	putObject  func(partName string, data []byte) error
}

func (zw *partWriter) Write(p []byte) (int, error) {
	const op errors.Op = "minio.partWriter.Write"
	if zw.err != nil {
		return 0, errors.E(op, zw.err)
	} else if len(p) == 0 {
		return 0, nil
	}

	partName := fmt.Sprintf("parts/%s/%s/%d", zw.mod, zw.ver, zw.numParts)

	zw.err = zw.putObject(partName, p)
	if zw.err != nil {
		return 0, errors.E(op, zw.err)
	}

	zw.srcs = append(zw.srcs, minio.CopySrcOptions{Bucket: zw.bucketName, Object: partName})
	zw.numParts++

	return len(p), nil
}

func (s *storageImpl) saveZip(ctx context.Context, dir, mod, ver string, zip io.Reader) error {
	const (
		op       errors.Op = "minio.saveZip"
		partSize           = 8 * 1024 * 1024
	)

	rdr := bufio.NewReaderSize(zip, partSize)
	wr := &partWriter{
		c:          s.minioClient,
		mod:        mod,
		ver:        ver,
		bucketName: s.bucketName,
		putObject: func(partName string, data []byte) error {
			_, err := s.minioClient.PutObject(ctx, s.bucketName, partName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
			return err
		},
	}

	_, err := rdr.WriteTo(wr)
	if err != nil {
		return errors.E(op, err)
	}

	zipFileName := dir + "/" + "source.zip"
	dst := minio.CopyDestOptions{Bucket: s.bucketName, Object: zipFileName}

	_, err = s.minioClient.ComposeObject(ctx, dst, wr.srcs...)
	if err != nil {
		return errors.E(op, errors.E("minio.ComposeObject", err))
	}

	err = s.removeParts(ctx, mod, ver, wr.numParts)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (s *storageImpl) removeParts(ctx context.Context, mod, ver string, numParts int) error {
	const op errors.Op = "minio.removeParts"

	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)

		for i := range numParts {
			objectsCh <- minio.ObjectInfo{Key: fmt.Sprintf("parts/%s/%s/%d", mod, ver, i)}
		}
	}()

	var errs error
	for e := range s.minioClient.RemoveObjects(ctx, s.bucketName, objectsCh, minio.RemoveObjectsOptions{}) {
		errs = multierror.Append(errs, e.Err)
	}

	if errs != nil {
		return errors.E(op, errs)
	}

	return nil
}
