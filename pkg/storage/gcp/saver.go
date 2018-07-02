package gcp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gobuffalo/envy"
	multierror "github.com/hashicorp/go-multierror"
	"google.golang.org/api/option"
)

// Storage implements the (github.com/gomods/pkg/storage).Saver interface
type Storage struct {
	bucket *storage.BucketHandle
}

// New returns a new Storage instance authenticated using the provided
// ClientOptions. The bucket name to be used will be loaded from the
// environment variable ATHENS_STORAGE_GCP_BUCKET.
//
// The ClientOptions should provide permissions sufficient to read, write and
// delete objects in google cloud storage for your project.
func New(ctx context.Context, cred option.ClientOption) (*Storage, error) {
	client, err := storage.NewClient(ctx, cred)
	if err != nil {
		return nil, fmt.Errorf("could not create new client: %s", err)
	}
	bucketname, err := envy.MustGet("ATHENS_STORAGE_GCP_BUCKET")
	if err != nil {
		return nil, fmt.Errorf("could not load 'ATHENS_STORAGE_GCP_BUCKET': %s", err)
	}
	bkt := client.Bucket(bucketname)
	return &Storage{bucket: bkt}, nil
}

// Save uploads the modules .mod, .zip and .info files for a given version
func (s *Storage) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	errs := make(chan error, 3)
	// create a context that will time out after 300 seconds / 5 minutes
	ctxWT, cancelCTX := context.WithTimeout(ctx, 300*time.Second)
	defer cancelCTX()

	// dispatch go routine for each file to upload
	go upload(ctxWT, errs, s.bucket, module, version, "mod", bytes.NewReader(mod))
	go upload(ctxWT, errs, s.bucket, module, version, "zip", zip)
	go upload(ctxWT, errs, s.bucket, module, version, "info", bytes.NewReader(info))

	var errors error
	// wait for each routine above to send a value
	for count := 0; count < 3; count++ {
		err := <-errs
		if err != nil {
			errors = multierror.Append(errors, err)
		}
	}
	close(errs)

	return errors
}

// upload waits for either writeToBucket to complete or the context expires
func upload(ctx context.Context, errs chan<- error, bkt *storage.BucketHandle, module, version, ext string, file io.Reader) {
	select {
	case errs <- writeToBucket(ctx, bkt, fmt.Sprintf("%s/@v/%s.%s", module, version, ext), file):
		return
	case <-ctx.Done():
		errs <- fmt.Errorf("WARNING: context deadline exceeded during write of %s version %s", module, version)
	}
}

// writeToBucket performs the actual write to a gcp storage bucket
func writeToBucket(ctx context.Context, bkt *storage.BucketHandle, filename string, file io.Reader) error {
	wc := bkt.Object(filename).NewWriter(ctx)
	defer func(w *storage.Writer) {
		if err := w.Close(); err != nil {
			log.Printf("WARNING: failed to close storage object writer: %v", err)
		}
	}(wc)
	wc.ContentType = "application/octet-stream"
	// TODO: set better access control?
	wc.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	if _, err := io.Copy(wc, file); err != nil {
		return err
	}
	return nil
}
