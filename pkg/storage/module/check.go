package module

import (
	"context"
	"fmt"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	multierror "github.com/hashicorp/go-multierror"
)

// Checker func takes a path to a module@version bit mod/info/zip and checks if it's present in the storage
type Checker func(ctx context.Context, path string) (bool, error)

// Exists checks if a mod@ver exists in the storage. If one or more of mod,info or zip bits is missing it returns false.
// Returns multierror containing errors from all checks and timeouts
func Exists(ctx context.Context, module, version string, exists Checker, timeout time.Duration) (bool, error) {
	const op errors.Op = "module.Exists"
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	type existsRes struct {
		exists bool
		err    error
	}
	check := func(ext string) <-chan *existsRes {
		ec := make(chan *existsRes)

		go func() {
			defer close(ec)
			p := config.PackageVersionedName(module, version, ext)
			exists, err := exists(tctx, p)
			ec <- &existsRes{exists, err}
		}()
		return ec
	}

	resChan := make(chan *existsRes, numFiles)
	checkOrAbort := func(ext string) {
		select {
		case res := <-check(ext):
			resChan <- res
		case <-tctx.Done():
			resChan <- &existsRes{false, fmt.Errorf("checking %s.%s.%s failed: %s", module, version, ext, tctx.Err())}
		}
	}

	go checkOrAbort("info")
	go checkOrAbort("mod")
	go checkOrAbort("zip")

	moduleExists := true
	var errs *multierror.Error

	for i := 0; i < numFiles; i++ {
		res := <-resChan
		if res.err != nil {
			errs = multierror.Append(errs, res.err)
		}
		if !res.exists {
			moduleExists = false
		}
	}
	close(resChan)

	if accError := errs.ErrorOrNil(); accError != nil {
		return false, errors.E(op, accError, errors.M(module), errors.V(version))
	}

	return moduleExists, nil
}
