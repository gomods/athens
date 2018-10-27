package multi

import (
	"context"
	"io"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

// Info implements the (./pkg/storage).Getter interface
func (s *Storage) Info(ctx context.Context, module, version string) ([]byte, error) {
	const op errors.Op = "mutli.Info"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	type FnResult struct {
		FileContent []byte
		Error       error
	}

	var results = make(chan FnResult, len(s.storages))
	c, cancel := context.WithCancel(ctx)

	for _, store := range s.storages {
		go func(sb storage.Backend) {
			r, e := sb.Info(ctx, module, version)
			select {
			case results <- FnResult{r, e}:
			case <-c.Done():
			}
		}(store)
	}

	var errs []error
	var result []byte
	for i := 0; i < len(s.storages); i++ {
		r := <-results
		if r.Error == nil {
			result = r.FileContent
			errs = nil
			break
		}

		errs = append(errs, r.Error)
	}

	cancel()
	close(results)

	return result, s.composeError(module, version, op, errs...)
}

// GoMod implements the (./pkg/storage).Getter interface
func (s *Storage) GoMod(ctx context.Context, module, version string) ([]byte, error) {
	const op errors.Op = "mutli.GoMod"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	type FnResult struct {
		FileContent []byte
		Error       error
	}

	var results = make(chan FnResult, len(s.storages))
	c, cancel := context.WithCancel(ctx)

	for _, store := range s.storages {
		go func(sb storage.Backend) {
			r, e := sb.GoMod(ctx, module, version)
			select {
			case results <- FnResult{r, e}:
			case <-c.Done():
			}
		}(store)
	}

	var errs []error
	var result []byte
	for i := 0; i < len(s.storages); i++ {
		r := <-results
		if r.Error == nil {
			result = r.FileContent
			errs = nil
			break
		}

		errs = append(errs, r.Error)
	}

	cancel()
	close(results)

	return result, s.composeError(module, version, op, errs...)
}

// Zip implements the (./pkg/storage).Getter interface
func (s *Storage) Zip(ctx context.Context, module, version string) (io.ReadCloser, error) {
	const op errors.Op = "mutli.Zip"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	type FnResult struct {
		ZipReadCloser io.ReadCloser
		Error         error
	}

	var results = make(chan FnResult, len(s.storages))
	c, cancel := context.WithCancel(ctx)

	for _, store := range s.storages {
		go func(sb storage.Backend) {
			r, e := sb.Zip(c, module, version)
			select {
			case results <- FnResult{r, e}:
				select {
				case <-c.Done():
					// in case cancellation required,
					// client has the result so we can close this one
					r.Close()
				default:
				}
			case <-c.Done():
			}
		}(store)
	}

	var errs []error
	var result io.ReadCloser
	for i := 0; i < len(s.storages); i++ {
		r := <-results
		if r.Error == nil {
			result = r.ZipReadCloser
			errs = nil
			break
		}

		errs = append(errs, r.Error)
	}

	cancel()
	close(results)

	return result, s.composeError(module, version, op, errs...)
}
