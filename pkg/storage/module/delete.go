package module

import (
	"context"
	"fmt"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/hashicorp/go-multierror"
)

// Deleter takes a path to a file and deletes it from the blob store.
type Deleter func(ctx context.Context, path string) error

// Delete deletes .info, .mod and .zip files from the blob store in parallel.
// Returns multierror containing errors from all deletes and timeouts.
func Delete(ctx context.Context, module, version string, del Deleter, timeout time.Duration) error {
	const op errors.Op = "module.Delete"
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	delFn := func(ext string) <-chan error {
		ec := make(chan error)

		go func() {
			defer close(ec)
			p := config.PackageVersionedName(module, version, ext)
			ec <- del(tctx, p)
		}()
		return ec
	}

	errChan := make(chan error, numFiles)
	delOrAbort := func(ext string) {
		select {
		case err := <-delFn(ext):
			errChan <- err
		case <-tctx.Done():
			errChan <- fmt.Errorf("deleting %s.%s.%s failed: %w", module, version, ext, tctx.Err())
		}
	}

	go delOrAbort("info")
	go delOrAbort("mod")
	go delOrAbort("zip")

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
