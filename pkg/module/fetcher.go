package module

import (
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

// Fetcher fetches module from an upstream source
type Fetcher interface {
	// Fetch downloads the sources from an upstream and returns the corresponding
	// .info, .mod, and .zip files.
	Fetch(ctx observ.ProxyContext, mod, ver string) (*storage.Version, error)
}
