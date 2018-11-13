package storage

import "context"

type ModVer struct {
	Module  string
	Version string
}

// Lister is the interface that lists versions of a specific baseURL & module
type Cataloger interface {
	// List gets all the versions for the given baseURL & module.
	// It returns ErrNotFound if the module isn't found
	Catalog(ctx context.Context, token string, elements int) ([]ModVer, string, error)
}
