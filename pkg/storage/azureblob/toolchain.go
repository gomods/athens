package azureblob

import (
	"context"
	"io"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/internal/toolchain"
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
	r, err := o.store.client.ReadBlob(ctx, key)
	if err != nil {
		return nil, releaseError(err)
	}
	return r, nil
}

func (o *releaseObjects) Create(ctx context.Context, key string, body io.ReadSeeker, _ int64) error {
	blob := o.store.client.containerURL.NewBlockBlobURL(key)
	_, err := azblob.UploadStreamToBlockBlob(ctx, body, blob, azblob.UploadStreamToBlockBlobOptions{BufferSize: 1 << 20, MaxBuffers: 3, AccessConditions: azblob.BlobAccessConditions{ModifiedAccessConditions: azblob.ModifiedAccessConditions{IfNoneMatch: azblob.ETagAny}}})
	return releaseError(err)
}

func (o *releaseObjects) Keys(ctx context.Context, prefix string) ([]string, error) {
	return o.store.client.ListBlobs(ctx, prefix)
}

func (o *releaseObjects) Delete(ctx context.Context, key string) error {
	return releaseError(o.store.client.DeleteBlob(ctx, key))
}

func releaseError(err error) error {
	if err == nil {
		return nil
	}
	var e azblob.StorageError
	if errors.AsErr(err, &e) {
		switch e.ServiceCode() {
		case azblob.ServiceCodeBlobNotFound, azblob.ServiceCodeContainerNotFound:
			return errors.E("azureblob.Toolchain", err, errors.KindNotFound)
		case azblob.ServiceCodeConditionNotMet:
			return errors.E("azureblob.Toolchain", err, errors.KindAlreadyExists)
		}
	}
	return err
}
