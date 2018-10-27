package multi

import (
	"context"

	"github.com/gomods/athens/pkg/storage"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// Exists implements the (./pkg/storage).Checker interface
// returning true if the module at version exists in storage
func (s *Storage) Exists(ctx context.Context, module, version string) (bool, error) {
	const op errors.Op = "mutli.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	type FnResult struct {
		Exists bool
		Error  error
	}

	var results = make(chan FnResult, len(s.storages))
	c, cancel := context.WithCancel(ctx)

	for _, store := range s.storages {
		go func(sb storage.Backend) {
			r, e := sb.Exists(c, module, version)
			select {
			case results <- FnResult{r, e}:
			case <-c.Done():
			}
		}(store)
	}

	var errs []error
	var result bool
	for i := 0; i < len(s.storages); i++ {
		r := <-results
		if r.Error == nil {
			result = r.Exists
			errs = nil
			break
		}

		errs = append(errs, r.Error)
	}

	cancel()
	close(results)

	return result, s.composeError(module, version, op, errs...)
}
