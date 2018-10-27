package multi

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

// Delete implements the (./pkg/storage).Deleter interface and
// removes a version of a module from storage. Returning ErrNotFound
// if the version does not exist.
func (s *Storage) Delete(ctx context.Context, module, version string) error {
	const op errors.Op = "mutli.Delete"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	var results = make(chan error, len(s.storages))
	c, cancel := context.WithCancel(ctx)

	for _, store := range s.storages {
		go func(sb storage.Backend) {
			select {
			case results <- sb.Delete(c, module, version):
			case <-c.Done():
			}
		}(store)
	}

	var errs []error
	for i := 0; i < len(s.storages); i++ {
		r := <-results
		if r != nil {
			errs = append(errs, r)
		}
	}

	cancel()
	close(results)

	return s.composeError(module, version, op, errs...)
}
