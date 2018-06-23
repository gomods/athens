package gcp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// Storage implements the (github.com/gomods/pkg/storage).Saver interface
type Storage struct {
	bucket *storage.BucketHandle
}

// New returns a new Storage instance
// authenticated using the provided ClientOptions for the associated bucket
func New(ctx context.Context, bucketname string, cred option.ClientOption) (*Storage, error) {
	client, err := storage.NewClient(ctx, cred)
	if err != nil {
		return nil, fmt.Errorf("could not create new client: %s", err)
	}
	bkt := client.Bucket(bucketname)
	return &Storage{bucket: bkt}, nil
}

// Save uploads the modules .mod, .zip and .info files for a given version
func (s *Storage) Save(ctx context.Context, module, version string, mod, zip, info []byte) error {
	errs := make(chan error, 3)
	// dispatch go routine for each file to upload
	go save(ctx, errs, s.bucket, module, version, "mod", mod)
	go save(ctx, errs, s.bucket, module, version, "zip", zip)
	go save(ctx, errs, s.bucket, module, version, "info", info)

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
func save(ctx context.Context, errs chan<- error, bkt *storage.BucketHandle, module, version, ext string, content []byte) {
	ctxWT, cancel := context.WithTimeout(ctx, 5*time.Minute)

	select {
	case errs <- writeToBucket(ctxWT, bkt, fmt.Sprintf("%s/@v/%s.%s", module, version, ext), content):
		cancel()
	case <-ctxWT.Done():
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
