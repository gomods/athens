package storage

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
)

// Checker is the interface that checks if the version of the module exists.
type Checker interface {
	// Exists checks whether or not module in specified version is present
	// in the backing storage.
	Exists(ctx context.Context, module, version string) (bool, error)
}

// WithChecker wraps the backend with a Checker implementation.
func WithChecker(strg Backend) Checker {
	if checker, ok := strg.(Checker); ok {
		return checker
	}
	return &checker{strg}
}

type checker struct {
	strg Backend
}

func (c *checker) Exists(ctx context.Context, module, version string) (bool, error) {
	_, err := c.strg.Info(ctx, module, version)
	if err != nil {
		if errors.Is(err, errors.KindNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
