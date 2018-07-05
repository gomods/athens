package gcp

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
	"github.com/gobuffalo/envy"
	"github.com/gomods/athens/pkg/config/env"
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

	datastore, err := datastore.NewClient(ctx, "", cred)
	if err != nil {
		return nil, fmt.Errorf("could not create new datastore client: %s", err)
	}
	return &Storage{
		bucket:       bkt,
		dsClient:     datastore,
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

// Save uploads the modules .mod, .zip and .info files for a given version
func (s *Storage) Save(ctx context.Context, module, version string, mod, zip, info []byte) error {
	errs := make(chan error, 3)
	// create a context that will time out after the value found in
	// the ATHENS_TIMEOUT env variable
	ctxWT, cancelCTX := context.WithTimeout(ctx, env.Timeout())
	defer cancelCTX()

	// dispatch go routine for each file to upload
	go upload(ctxWT, errs, s.bucket, module, version, "mod", mod)
	go upload(ctxWT, errs, s.bucket, module, version, "zip", zip)
	go upload(ctxWT, errs, s.bucket, module, version, "info", info)

	errsOut := make([]string, 0, 3)
	// wait for each routine above to send a value
	for count := 0; count < 3; count++ {
		err := <-errs
		if err != nil {
			errsOut = append(errsOut, err.Error())
		}
	}
	close(errs)

	// return concatenated error string if there is anything to report
	if len(errsOut) > 0 {
		return fmt.Errorf("one or more errors occured saving %s/@v/%s: %s", module, version, strings.Join(errsOut, ", "))
	}
	return nil
}

// upload waits for either writeToBucket to complete or the context to expire
func upload(ctx context.Context, errs chan<- error, bkt *storage.BucketHandle, module, version, ext string, file []byte) {
	select {
	case errs <- writeToBucket(ctx, bkt, fmt.Sprintf("%s/@v/%s.%s", module, version, ext), file):
		return
	case <-ctx.Done():
		errs <- fmt.Errorf("WARNING: context deadline exceeded during write of %s version %s", module, version)
	}
}

// writeToBucket performs the actual write to a gcp storage bucket
func writeToBucket(ctx context.Context, bkt *storage.BucketHandle, filename string, file []byte) error {
	wc := bkt.Object(filename).NewWriter(ctx)
	defer func(w *storage.Writer) {
		if err := w.Close(); err != nil {
			log.Printf("WARNING: failed to close storage object writer: %v", err)
		}
	}(wc)
	wc.ContentType = "application/octet-stream"
	// TODO: set better access control?
	wc.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	if _, err := wc.Write(file); err != nil {
		return err
	}
	return nil
}
