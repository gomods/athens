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
	minio "github.com/minio/minio-go/v6"
)

func (s *storageImpl) Save(ctx context.Context, module, vsn string, mod []byte, zip io.Reader, info []byte) error {
	const op errors.Op = "storage.minio.Save"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	dir := s.versionLocation(module, vsn)
	modFileName := dir + "/" + "go.mod"
	infoFileName := dir + "/" + vsn + ".info"
	_, err := s.minioClient.PutObject(s.bucketName, modFileName, bytes.NewReader(mod), int64(len(mod)), minio.PutObjectOptions{})
	if err != nil {
		return errors.E(op, err)
	}
	// Chunk the stream into 8mb and send them in parts to minio.
	// This is because the minio client over-allocates a stream buffer (600Mb)
	// when the size is unknown, see https://github.com/minio/minio-go/issues/848
	err = s.saveZip(dir, module, vsn, zip)
	if err != nil {
		return errors.E(op, err)
	}
	_, err = s.minioClient.PutObject(s.bucketName, infoFileName, bytes.NewReader(info), int64(len(info)), minio.PutObjectOptions{})
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
	srcs       []minio.SourceInfo
}

func (zw *partWriter) Write(p []byte) (int, error) {
	const op errors.Op = "minio.partWriter.Write"
	if zw.err != nil {
		return 0, errors.E(op, zw.err)
	} else if len(p) == 0 {
		return 0, nil
	}
	partName := fmt.Sprintf("parts/%s/%s/%d", zw.mod, zw.ver, zw.numParts)
	plen := int64(len(p))
	_, zw.err = zw.c.PutObject(zw.bucketName, partName, bytes.NewReader(p), plen, minio.PutObjectOptions{})
	if zw.err != nil {
		return 0, errors.E(op, zw.err)
	}
	zw.srcs = append(zw.srcs, minio.NewSourceInfo(zw.bucketName, partName, nil))
	zw.numParts++
	return len(p), nil
}

func (s *storageImpl) saveZip(dir, mod, ver string, zip io.Reader) error {
	const op errors.Op = "minio.saveZip"
	const partSize = 8 * 1024 * 1024
	rdr := bufio.NewReaderSize(zip, partSize)
	wr := &partWriter{0, s.minioClient, mod, ver, s.bucketName, nil, nil}
	_, err := rdr.WriteTo(wr)
	if err != nil {
		return errors.E(op, err)
	}
	zipFileName := dir + "/" + "source.zip"
	dst, err := minio.NewDestinationInfo(s.bucketName, zipFileName, nil, nil)
	if err != nil {
		return errors.E(op, errors.E("minio.NewDestinationInfo", err))
	}
	err = s.minioClient.ComposeObject(dst, wr.srcs)
	if err != nil {
		return errors.E(op, errors.E("minio.ComposeObject", err))
	}
	err = s.removeParts(mod, ver, wr.numParts)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (s *storageImpl) removeParts(mod, ver string, numParts int) error {
	const op errors.Op = "minio.removeParts"
	objectsCh := make(chan string)
	go func() {
		defer close(objectsCh)
		for i := 0; i < numParts; i++ {
			objectsCh <- fmt.Sprintf("parts/%s/%s/%d", mod, ver, i)
		}
	}()
	var errs error
	for e := range s.minioClient.RemoveObjects(s.bucketName, objectsCh) {
		errs = multierror.Append(errs, e.Err)
	}
	if errs != nil {
		return errors.E(op, errs)
	}
	return nil
}
