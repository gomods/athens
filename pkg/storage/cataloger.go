package storage

import (
	"context"

	"github.com/gomods/athens/pkg/paths"
)

// Cataloger is the interface that lists all the modules and version contained in the storage
type Cataloger interface {
	// Catalog gets all the modules / versions.
	// It returns ErrNotFound if the module isn't found
	Catalog(ctx context.Context, token string, elements int) ([]paths.AllPathParams, string, error)
}
