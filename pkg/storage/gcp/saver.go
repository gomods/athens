package gcp

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/gobuffalo/envy"
	"github.com/gomods/athens/pkg/config/env"
	"google.golang.org/api/option"
	"google.golang.org/appengine/file"
)

// Storage implements the (github.com/gomods/pkg/storage).Saver interface
type Storage struct {
	bucket  *storage.BucketHandle
	baseURI *url.URL
	close   func() error
}

// NewWithCredentials returns a new Storage instance authenticated using the provided
// ClientOptions. The bucket name to be used will be loaded from the
// environment variable ATHENS_STORAGE_GCP_BUCKET.
//
// The ClientOptions should provide permissions sufficient to read, write and
// delete objects in google cloud storage for your project.
func NewWithCredentials(ctx context.Context, cred option.ClientOption) (*Storage, error) {
	client, err := storage.NewClient(ctx, cred)
	if err != nil {
		return nil, fmt.Errorf("could not create new client: %s", err)
	}
	bucketname, err := envy.MustGet("ATHENS_STORAGE_GCP_BUCKET")
	if err != nil {
		return nil, fmt.Errorf("could not load 'ATHENS_STORAGE_GCP_BUCKET': %s", err)
	}
	u, err := url.Parse(fmt.Sprintf("https://storage.googleapis.com/%s", bucketname))
	if err != nil {
		return nil, err
	}
	bkt := client.Bucket(bucketname)
	return &Storage{bucket: bkt, baseURI: u, close: client.Close}, nil
}

// New returns a new Storage instance for use with appengine hosts.
//
// This requires the application to be running on appengine, but does not require
// any special configuration.
// If running on any other platform please use NewWithCredentials
func New(ctx context.Context) (*Storage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create new client: %s", err)
	}
	bucketname, err := file.DefaultBucketName(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get default bucket name: %s", err)
	}
	u, err := url.Parse(fmt.Sprintf("https://storage.googleapis.com/%s", bucketname))
	if err != nil {
		return nil, err
	}
	bkt := client.Bucket(bucketname)
	return &Storage{bucket: bkt, baseURI: u, close: client.Close}, nil
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

// Close calls the underlying storage client's close method
//
// Close need not be called at program exit, it is provided in case a need arises.
func (s *Storage) Close() error {
	return s.close()
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
	case errs <- writeToBucket(ctx, bkt, key(module, version, ext), file):
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

// key returns the fully formatted module string
// eg:
//		gomods/athens/@v/v1.2.3.info
func key(module, version, filetype string) string {
	return fmt.Sprintf("%s/@v/%s.%s", module, version, filetype)
}
