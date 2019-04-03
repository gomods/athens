package gcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// Storage implements the (./pkg/storage).Backend interface
type Storage struct {
	bucket  *storage.BucketHandle
	timeout time.Duration
}

// NewClient creates a new storage client, without making any network calls.
// Most callers will want to call New, because that does everything needed
// to create a new Storage implementation that is backed by GCP. This function
// is helpful primarily for tests.
//
// Here's why: test code needs to create a GCP storage client _without_ doing
// any network calls, since it needs to create a unique bucket for testing
// purposes. Since the below call to 'New' does some validation on the bucket,
// they need to call this function, create the new bucket, and then call New
func NewClient(ctx context.Context, jsonKey string) (*storage.Client, error) {
	const op errors.Op = "gcp.newClient"

	opts := []option.ClientOption{}
	if jsonKey != "" {
		key, err := base64.StdEncoding.DecodeString(jsonKey)
		if err != nil {
			return nil, errors.E(op, fmt.Errorf("could not decode base64 json key: %v", err))
		}
		creds, err := google.CredentialsFromJSON(ctx, key, storage.ScopeReadWrite)
		if err != nil {
			return nil, errors.E(op, fmt.Errorf("could not get GCS credentials: %v", err))
		}
		opts = append(opts, option.WithCredentials(creds))
	}
	s, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("could not create new storage client: %s", err))
	}
	return s, nil
}

// New returns a new Storage instance backed by a Google Cloud Storage bucket.
// The bucket name to be used will be gcpConf.Bucket. Most commonly, this name
// will be loaded from the ATHENS_STORAGE_GCP_BUCKET environment variable.
//
// If you're not running on GCP, set the GOOGLE_APPLICATION_CREDENTIALS environment variable
// to the path of your service account file. If you're running on GCP (e.g. AppEngine),
// credentials will be automatically provided.
// See https://cloud.google.com/docs/authentication/getting-started.
func New(ctx context.Context, gcpConf *config.GCPConfig, timeout time.Duration) (*Storage, error) {
	const op errors.Op = "gcp.New"
	s, err := NewClient(ctx, gcpConf.JSONKey)
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
		bucket:  bkt,
		timeout: timeout,
	}, nil
}
