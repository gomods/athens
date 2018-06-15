package gcp

import (
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/bketelsen/buffet"
	"github.com/gobuffalo/buffalo"
	"google.golang.org/api/option"
)

// Storage implements the (github.com/gomods/pkg/storage).Saver interface
type Storage struct {
	bucket *storage.BucketHandle
}

// New returns a new Storage instance
// authenticated using the provided ClientOptions
func New(ctx buffalo.Context, cred option.ClientOption) (*Storage, error) {
	client, err := storage.NewClient(ctx, cred)
	if err != nil {
		return nil, fmt.Errorf("could not create new client: %s", err)
	}
	// The bucket MUST already exist
	bkt := client.Bucket("gomodules")
	return &Storage{bucket: bkt}, nil
}

// Save uploads the modules .mod, .zip and .info files for a given version
func (s *Storage) Save(ctx buffalo.Context, module, version string, mod, zip, info []byte) error {
	sp := buffet.ChildSpan("storage.save", ctx)
	defer sp.Finish()

	errs := make(chan error, 3)
	// dispatch go routine for each file to upload
	// TODO: make this compact
	go func(errs chan<- error) {
		sp := buffet.ChildSpan("storage.save.module", ctx)
		defer sp.Finish()

		if err := save(ctx, s.bucket, fmt.Sprintf("%s/@v/%s.mod", module, version), mod); err != nil {
			errs <- err
		} else {
			errs <- nil
		}
	}(errs)
	go func(errs chan<- error) {
		sp := buffet.ChildSpan("storage.save.zip", ctx)
		defer sp.Finish()

		if err := save(ctx, s.bucket, fmt.Sprintf("%s/@v/%s.zip", module, version), zip); err != nil {
			errs <- err
		} else {
			errs <- nil
		}
	}(errs)
	go func(errs chan<- error) {
		sp := buffet.ChildSpan("storage.save.info", ctx)
		defer sp.Finish()

		if err := save(ctx, s.bucket, fmt.Sprintf("%s/@v/%s.info", module, version), info); err != nil {
			errs <- err
		} else {
			errs <- nil
		}
	}(errs)

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

// save performs the actual write to a gcp storage bucket
func save(ctx buffalo.Context, bkt *storage.BucketHandle, filename string, file []byte) error {
	sp := buffet.ChildSpan("storage.save.writer", ctx)
	defer sp.Finish()

	wc := bkt.Object(filename).NewWriter(ctx)
	wc.ContentType = "application/octet-stream"
	// TODO: set better access control?
	wc.ACL = []storage.ACLRule{{storage.AllUsers, storage.RoleReader}}
	if _, err := wc.Write(file); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return nil
}
