package module

import (
	"context"
	"fmt"

	"github.com/gomods/athens/pkg/config"
	multierror "github.com/hashicorp/go-multierror"
)

// Deleter takes a path to a file and deletes it from the blob store
type Deleter func(ctx context.Context, path string) error

// Delete deletes .info, .mod and .zip files from the blob store in parallel.
// Returns multierror containing errors from all deletes and timeouts
func Delete(ctx context.Context, module, version string, delete Deleter) error {
	errChan := make(chan error, numFiles)

	del := func(ext string) {
		p := config.PackageVersionedName(module, version, ext)
		select {
		case errChan <- delete(ctx, p):
		case <-ctx.Done():
			errChan <- fmt.Errorf("deleting %s timed out", p)
		}
	}

	go del("info")
	go del("mod")
	go del("zip")

	var errors error
	for i := 0; i < numFiles; i++ {
		err := <-errChan
		if err != nil {
			errors = multierror.Append(errors, err)
		}
	}
	close(errChan)

	return errors
}
