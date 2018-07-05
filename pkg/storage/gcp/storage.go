package gcp

import (
	"context"
	"fmt"
	"net/url"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
	"github.com/gobuffalo/envy"
	multierror "github.com/hashicorp/go-multierror"
	"google.golang.org/api/option"
)

// Storage implements the (github.com/gomods/pkg/storage).Saver interface
type Storage struct {
	bucket       *storage.BucketHandle
	dsClient     *datastore.Client
	baseURI      *url.URL
	closeStorage func() error
}

// NewWithCredentials returns a new Storage instance authenticated using the provided
// ClientOptions. The bucket name to be used will be loaded from the
// environment variable ATHENS_STORAGE_GCP_BUCKET.
// TODO: project ID for datastore
//
// The ClientOptions should provide permissions sufficient to read, write and
// delete objects in google cloud storage for your project.
// TODO: As well as datastore
func NewWithCredentials(ctx context.Context, cred option.ClientOption) (*Storage, error) {
	storage, err := storage.NewClient(ctx, cred)
	if err != nil {
		return nil, fmt.Errorf("could not create new storage client: %s", err)
	}
	bucketname, err := envy.MustGet("ATHENS_STORAGE_GCP_BUCKET")
	if err != nil {
		return nil, fmt.Errorf("could not load 'ATHENS_STORAGE_GCP_BUCKET': %s", err)
	}
	u, err := url.Parse(fmt.Sprintf("https://storage.googleapis.com/%s", bucketname))
	if err != nil {
		return nil, err
	}
	bkt := storage.Bucket(bucketname)

	// datastore, err := datastore.NewClient(ctx, "", cred)
	// if err != nil {
	// 	return nil, fmt.Errorf("could not create new datastore client: %s", err)
	// }
	return &Storage{
		bucket: bkt,
		// dsClient:     datastore,
		baseURI:      u,
		closeStorage: storage.Close,
	}, nil
}

// BaseURL returns the base URL that stores all modules. It can be used
// in the "meta" tag redirect response to vgo.
//
// For example:
//
//	<meta name="go-import" content="gomods.com/athens mod BaseURL()">
func (s *Storage) BaseURL() *url.URL {
	return s.baseURI
}

// Close calls the underlying storage and datastore client's close methods
func (s *Storage) Close() error {
	var errors error
	if err := s.closeStorage(); err != nil {
		errors = multierror.Append(errors, err)
	}
	if err := s.dsClient.Close(); err != nil {
		errors = multierror.Append(errors, err)
	}
	return errors
}
