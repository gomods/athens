package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
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
	out, err := o.store.s3API.GetObject(ctx, &awss3.GetObjectInput{Bucket: aws.String(o.store.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, releaseError(err)
	}
	return storage.NewSizer(out.Body, aws.ToInt64(out.ContentLength)), nil
}

func (o *releaseObjects) Create(ctx context.Context, key string, body io.ReadSeeker, size int64) error {
	s := o.store
	_, err := s.s3API.PutObject(ctx, &awss3.PutObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key), Body: body, ContentLength: aws.Int64(size), ContentType: aws.String("application/octet-stream"), IfNoneMatch: aws.String("*"), ServerSideEncryption: types.ServerSideEncryption(s.serverSideEncryption), SSEKMSKeyId: func() *string {
		if s.sseKMSKeyID == "" {
			return nil
		}
		return aws.String(s.sseKMSKeyID)
	}(), BucketKeyEnabled: s.bucketKeyEnabled})
	return releaseError(err)
}

func (o *releaseObjects) Keys(ctx context.Context, prefix string) ([]string, error) {
	p := awss3.NewListObjectsV2Paginator(o.store.s3API, &awss3.ListObjectsV2Input{Bucket: aws.String(o.store.bucket), Prefix: aws.String(prefix)})
	var keys []string
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, releaseError(err)
		}
		for _, obj := range page.Contents {
			keys = append(keys, aws.ToString(obj.Key))
		}
	}
	return keys, nil
}

func (o *releaseObjects) Delete(ctx context.Context, key string) error {
	_, err := o.store.s3API.HeadObject(ctx, &awss3.HeadObjectInput{Bucket: aws.String(o.store.bucket), Key: aws.String(key)})
	if err != nil {
		return releaseError(err)
	}
	_, err = o.store.s3API.DeleteObject(ctx, &awss3.DeleteObjectInput{Bucket: aws.String(o.store.bucket), Key: aws.String(key)})
	return releaseError(err)
}

func releaseError(err error) error {
	if err == nil {
		return nil
	}
	var api smithy.APIError
	if errors.AsErr(err, &api) {
		switch api.ErrorCode() {
		case "NoSuchKey", "NotFound":
			return errors.E("s3.Toolchain", err, errors.KindNotFound)
		case "PreconditionFailed", "ConditionalRequestConflict":
			return errors.E("s3.Toolchain", err, errors.KindAlreadyExists)
		}
	}
	return errors.E("s3.Toolchain", err)
}
