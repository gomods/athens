package storage

import "github.com/gomods/athens/pkg/observ"

// Lister is the interface that lists versions of a specific baseURL & module
type Lister interface {
	// List gets all the versions for the given baseURL & module.
	// It returns ErrNotFound if the module isn't found
	List(ctx observ.ProxyContext, module string) ([]string, error)
}
