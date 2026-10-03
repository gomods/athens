package gcp

import (
	"context"
	"io"

	cloudstorage "cloud.google.com/go/storage"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/internal/toolchain"
	"google.golang.org/api/iterator"
)

func (s *Storage) toolchains() *toolchain.Store { return &toolchain.Store{Objects: &releaseObjects{s}} }
func (s *Storage) Archive(ctx context.Context, source, filename string) (storage.ArchiveInfo, storage.SizeReadCloser, error) {
	return s.toolchains().Archive(ctx, source, filename)
}

func (s *Storage) SaveArchive(ctx context.Context, source string, info storage.ArchiveInfo, body io.Reader) error {
	return s.toolchains().SaveArchive(ctx, source, info, body)
}

func (s *Storage) ListArchives(ctx context.Context, source string) ([]storage.ArchiveInfo, error) {
	return s.toolchains().ListArchives(ctx, source)
}

func (s *Storage) DeleteArchive(ctx context.Context, source, filename string) error {
	return s.toolchains().DeleteArchive(ctx, source, filename)
}

func (s *Storage) Releases(ctx context.Context, source string) (*storage.ReleaseList, error) {
	return s.toolchains().Releases(ctx, source)
}

func (s *Storage) SaveReleases(ctx context.Context, source string, list *storage.ReleaseList) error {
	return s.toolchains().SaveReleases(ctx, source, list)
}

type releaseObjects struct{ store *Storage }

func (o *releaseObjects) Open(ctx context.Context, key string) (storage.SizeReadCloser, error) {
	r, err := o.store.bucket.Object(key).NewReader(ctx)
	if err != nil {
		return nil, errors.E("gcp.Toolchain", err, getErrorKind(err))
	}
	return storage.NewSizer(r, r.Attrs.Size), nil
}

func (o *releaseObjects) Create(ctx context.Context, key string, body io.ReadSeeker, _ int64) error {
	return o.store.upload(ctx, key, body, nil, false)
}

func (o *releaseObjects) Keys(ctx context.Context, prefix string) ([]string, error) {
	it := o.store.bucket.Objects(ctx, &cloudstorage.Query{Prefix: prefix})
	var keys []string
	for {
		attrs, err := it.Next()
		if errors.IsErr(err, iterator.Done) {
			return keys, nil
		}
		if err != nil {
			return nil, err
		}
		keys = append(keys, attrs.Name)
	}
}

func (o *releaseObjects) Delete(ctx context.Context, key string) error {
	err := o.store.bucket.Object(key).Delete(ctx)
	if err != nil {
		return errors.E("gcp.DeleteArchive", err, getErrorKind(err))
	}
	return nil
}
