package module

import (
	"context"

	"github.com/gomods/athens/pkg/storage"
)

// Fetcher fetches module from an upstream source
type Fetcher interface {
	// Fetch downloads the sources from an upstream and returns the corresponding
	// .info, .mod, and .zip files.
	Fetch(ctx context.Context, mod, ver string) (*storage.Version, error)
}
