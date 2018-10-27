package multi

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

// List implements the (./pkg/storage).Lister interface
// It returns a list of versions, if any, for a given module
func (s *Storage) List(ctx context.Context, module string) ([]string, error) {
	const op errors.Op = "mutli.List"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	type FnResult struct {
		Versions []string
		Error    error
	}

	var results = make(chan FnResult, len(s.storages))
	c, cancel := context.WithCancel(ctx)

	for _, store := range s.storages {
		go func(sb storage.Backend) {
			r, e := sb.List(c, module)
			select {
			case results <- FnResult{r, e}:
			case <-c.Done():
			}
		}(store)
	}

	var err error
	var result []string
	for i := 0; i < len(s.storages); i++ {
		r := <-results
		if r.Error != nil {
			err = r.Error
			break
		}

		result = append(result, r.Versions...)
	}

	cancel()
	close(results)

	return result, s.composeError(module, "", op, err)
}
