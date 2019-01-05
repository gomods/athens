package gcp

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"google.golang.org/api/iterator"
)

// Storage implements the (./pkg/storage).Backend interface
type Storage struct {
	bucket       *storage.BucketHandle
	closeStorage func() error
	projectID    string
	timeout      time.Duration
}

// New returns a new Storage instance backed by a Google Cloud Storage bucket.
// The bucket name to be used will be loaded from the
// environment variable ATHENS_STORAGE_GCP_BUCKET.
//
// If you're not running on GCP, set the GOOGLE_APPLICATION_CREDENTIALS environment variable
// to the path of your service account file. If you're running on GCP (e.g. AppEngine),
// credentials will be automatically provided.
// See https://cloud.google.com/docs/authentication/getting-started.
func New(ctx context.Context, gcpConf *config.GCPConfig, timeout time.Duration) (*Storage, error) {
	const op errors.Op = "gcp.New"
	s, err := storage.NewClient(ctx)
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("could not create new storage client: %s", err))
	}

	bkt := s.Bucket(gcpConf.Bucket)
	if _, err := bkt.Attrs(ctx); err != nil {
		if err == storage.ErrBucketNotExist {
			return nil, errors.E(op, "You must manually create a storage bucket for Athens, see https://cloud.google.com/storage/docs/creating-buckets#storage-create-bucket-console")
		}
		return nil, errors.E(op, err)
	}

	return &Storage{
		bucket:       bkt,
		closeStorage: s.Close,
		timeout:      timeout,
	}, nil
}

// Close calls the underlying storage client's close method
// It is not required to be called on program exit but provided here
// for completness.
func (s *Storage) Close() error {
	return s.closeStorage()
}

func (s *Storage) delete(ctx context.Context, path string) error {
	err := s.bucket.Object(path).Delete(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) open(ctx context.Context, path string) (io.ReadCloser, error) {
	rc, err := s.bucket.Object(path).NewReader(ctx)
	if err != nil {
		return rc, err
	}
	return rc, nil
}

func (s *Storage) write(ctx context.Context, path string) io.WriteCloser {
	return s.bucket.Object(path).NewWriter(ctx)
}

func (s *Storage) list(ctx context.Context, prefix string) ([]string, error) {
	it := s.bucket.Objects(ctx, &storage.Query{Prefix: prefix})

	res := []string{}
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		res = append(res, attrs.Name)
	}

	return res, nil
}

func (s *Storage) exists(ctx context.Context, path string) (bool, error) {
	_, err := s.bucket.Object(path).Attrs(ctx)

	if err == storage.ErrObjectNotExist {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *Storage) catalog(ctx context.Context, token string, pageSize int) ([]string, string, error) {
	it := s.bucket.Objects(ctx, nil)
	p := iterator.NewPager(it, pageSize, token)

	attrs := make([]*storage.ObjectAttrs, 0)
	nextToken, err := p.NextPage(&attrs)
	if err != nil {
		return nil, "", err
	}

	res := []string{}
	for _, attr := range attrs {
		res = append(res, attr.Name)
	}
	return res, nextToken, nil
}
