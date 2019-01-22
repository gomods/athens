package module

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	multierror "github.com/hashicorp/go-multierror"
)

const numFiles = 3

// Uploader takes a stream and saves it to the blob store under a given path
type Uploader func(ctx context.Context, path, contentType string, stream Stream) error

// Stream is the object that is passed along to Uploaders
type Stream struct {
	Stream io.Reader
	Size   int64
}

// NewStreamFromBytes returns a new module.Stream from a slice of bytes to Savers
func NewStreamFromBytes(b []byte) Stream {
	return Stream{
		Stream: bytes.NewReader(b),
		Size:   int64(len(b)),
	}
}

// NewStreamFromReaderWithSize returns a new module.Stream from an io.Reader and keeps the len -1 for minio storage > 600 MB
func NewStreamFromReaderWithSize(r io.Reader, s int64) Stream {
	const MinioLimit = 600000000
	if s > MinioLimit {
		s = -1
	}
	return Stream{
		Stream: r,
		Size:   s,
	}
}

// Upload saves .info, .mod and .zip files to the blob store in parallel.
// Returns multierror containing errors from all uploads and timeouts
func Upload(ctx context.Context, module, version string, info, mod, zip Stream, uploader Uploader, timeout time.Duration) error {
	const op errors.Op = "module.Upload"
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	save := func(ext, contentType string, stream Stream) <-chan error {
		ec := make(chan error)

		go func() {
			defer close(ec)
			p := config.PackageVersionedName(module, version, ext)
			ec <- uploader(tctx, p, contentType, stream)
		}()
		return ec
	}

	errChan := make(chan error, numFiles)
	saveOrAbort := func(ext, contentType string, stream Stream) {
		select {
		case err := <-save(ext, contentType, stream):
			errChan <- err
		case <-tctx.Done():
			errChan <- fmt.Errorf("uploading %s.%s.%s failed: %s", module, version, ext, tctx.Err())
		}
	}

	go saveOrAbort("info", "application/json", info)
	go saveOrAbort("mod", "text/plain", mod)
	go saveOrAbort("zip", "applicatlsion/octet-stream", zip)

	var errs error
	for i := 0; i < numFiles; i++ {
		err := <-errChan
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	close(errChan)
	if errs != nil {
		return errors.E(op, errs)
	}

	return nil
}
