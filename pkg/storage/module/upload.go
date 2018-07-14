package module

import (
	"context"
	"fmt"
	"io"

	"github.com/gomods/athens/pkg/config"
	multierror "github.com/hashicorp/go-multierror"
)

const numUpload = 3

// Uploader takes a stream and saves it to the blob store under a given path
type Uploader func(ctx context.Context, path, contentType string, stream io.Reader) error

// Upload saves .info, .mod and .zip files to the blob store in parallel.
// Returns multierror containing errors from all uploads and timeouts
func Upload(ctx context.Context, module, version string, info, mod, zip io.Reader, uploader Uploader) error {
	errChan := make(chan error, numUpload)

	save := func(ext, contentType string, stream io.Reader) {
		p := config.PackageVersionedName(module, version, ext)
		select {
		case errChan <- uploader(ctx, p, contentType, stream):
		case <-ctx.Done():
			errChan <- fmt.Errorf("uploading %s timed out", p)
		}
	}

	go save("info", "application/json", info)
	go save("mod", "text/plain", mod)
	go save("zip", "application/octet-stream", zip)

	var errors error
	for i := 0; i < numUpload; i++ {
		err := <-errChan
		if err != nil {
			errors = multierror.Append(errors, err)
		}
	}
	close(errChan)

	return errors
}
