package minio

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	minio "github.com/minio/minio-go"
	"golang.org/x/sync/errgroup"
)

type modMeta struct {
	file string
	len  int64
	data io.Reader
}

// Save takes in a module and its data and tries to save it in the minio bucket
func (s *storageImpl) Save(ctx context.Context, module, vsn string, mod []byte, zip io.Reader, info []byte) error {
	const op errors.Op = "storage.minio.Save"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	var eg errgroup.Group
	dir := s.versionLocation(module, vsn)

	mS := [3]modMeta{
		{file: dir + "/" + "go.mod", len: int64(len(mod)), data: bytes.NewReader(mod)},
		{file: dir + "/" + "source.zip", len: -1, data: zip},
		{file: dir + "/" + vsn + ".info", len: int64(len(info)), data: bytes.NewBuffer(info)},
	}

	for _, m := range mS {
		m := m
		eg.Go(func() error {
			_, err := s.minioClient.PutObject(s.bucketName, m.file, m.data, m.len, minio.PutObjectOptions{})
			return err
		})
	}

	if err := eg.Wait(); err != nil {
		// Best effort delete when any one of the save fails
		var wg sync.WaitGroup
		for _, m := range mS {
			wg.Add(1)
			go func(m modMeta) {
				_ = s.minioClient.RemoveObject(s.bucketName, m.file)
				wg.Done()
			}(m)
		}
		wg.Wait()
	}
	return nil
}
