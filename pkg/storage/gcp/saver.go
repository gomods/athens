package gcp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gobuffalo/envy"
	"google.golang.org/api/option"
)

// Storage implements the (github.com/gomods/pkg/storage).Saver interface
type Storage struct {
	bucket *storage.BucketHandle
}

// New returns a new Storage instance authenticated using the provided
// ClientOptions. The bucket name to be used will be loaded from the
// environment variable ATHENS_GCP_BUCKET_NAME.
//
// The ClientOptions should provide permissions sufficient to read, write and
// delete objects in google cloud storage for your project.
func New(ctx context.Context, cred option.ClientOption) (*Storage, error) {
	client, err := storage.NewClient(ctx, cred)
	if err != nil {
		return nil, fmt.Errorf("could not create new client: %s", err)
	}
	bucketname, err := envy.MustGet("ATHENS_GCP_BUCKET_NAME")
	if err != nil {
		return nil, fmt.Errorf("could not load 'ATHENS_GCP_BUCKET_NAME': %s", err)
	}
	bkt := client.Bucket(bucketname)
	return &Storage{bucket: bkt}, nil
}

// Save uploads the modules .mod, .zip and .info files for a given version
func (s *Storage) Save(ctx context.Context, module, version string, mod, zip, info []byte) error {
	errs := make(chan error, 3)
	// create a context that will time out after 300 seconds / 5 minutes
	ctxWT, cancel := context.WithTimeout(ctx, 300*time.Second)

	// dispatch go routine for each file to upload
	go save(ctxWT, cancel, errs, s.bucket, module, version, "mod", mod)
	go save(ctxWT, cancel, errs, s.bucket, module, version, "zip", zip)
	go save(ctxWT, cancel, errs, s.bucket, module, version, "info", info)

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

// save waits for writeToBucket to complete or times out after five minutes
func save(ctx context.Context, cancel func(), errs chan<- error, bkt *storage.BucketHandle, module, version, ext string, content []byte) {
	select {
	case errs <- writeToBucket(ctx, bkt, fmt.Sprintf("%s/@v/%s.%s", module, version, ext), content):
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
