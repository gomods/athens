package gcp

import (
	"context"
	"fmt"
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
func New(bucketname string, cred option.ClientOption) (*Storage, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, cred)
	if err != nil {
		return nil, fmt.Errorf("could not create new client: %s", err)
	}
	bkt := client.Bucket(bucketname)
	return &Storage{bucket: bkt}, nil
}

// Save uploads the modules .mod, .zip and .info files for a given version
func (s *Storage) Save(module, version string, mod, zip, info []byte) error {
	ctx := context.Background()
	errs := make(chan error, 3)
	// dispatch go routine for each file to upload
	go save(ctx, errs, s.bucket, module, version, "mod", mod)
	go save(ctx, errs, s.bucket, module, version, "zip", zip)
	go save(ctx, errs, s.bucket, module, version, "info", info)

	errsOut := make([]string, 3)
	// wait for each routine above to send a value
	for count := 0; count < 3; count++ {
		err := <-errs
		if err != nil {
			errsOut = append(errsOut, err.Error())
		}
	}

	// return concatenated error string if there is anything to report
	if len(errsOut) < 0 {
		return fmt.Errorf("one or more errors occured saving %s/@v/%s: %s", module, version, strings.Join(errsOut, ", "))
	}
	return nil
}

// save waits for writeToBucket to complete or times out after five minutes
func save(ctx context.Context, errs chan<- error, bkt *storage.BucketHandle, module, version, ext string, content []byte) {
	select {
	case errs <- writeToBucket(ctx, bkt, fmt.Sprintf("%s/@v/%s.%s", module, version, ext), content):
		return
	case <-time.After(5 * time.Minute):
		errs <- fmt.Errorf("WARNING: write of %s version %s timed out", module, version)
	}
}

// writeToBucket performs the actual write to a gcp storage bucket
func writeToBucket(ctx context.Context, bkt *storage.BucketHandle, filename string, file []byte) error {
	wc := bkt.Object(filename).NewWriter(ctx)
	wc.ContentType = "application/octet-stream"
	// TODO: set better access control?
	wc.ACL = []storage.ACLRule{{storage.AllUsers, storage.RoleReader}}
	if _, err := wc.Write(file); err != nil {
		return err
	}
	return wc.Close()
}
