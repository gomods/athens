package module

import (
	"github.com/gomods/athens/pkg/storage"
)

// Fetcher fetches module from an upstream source
type Fetcher interface {
	// Fetch fetches the module and puts it somewhere addressable by ModuleRef.
	// returns a non-nil error on failure.
	//
	// The caller should call moduleRef.Clear() after they're done with the module
	Fetch(mod, ver string) (Ref, error)
}

// Ref points to a module somewhere
type Ref interface {
	// Clear frees the storage & resources that the module uses
	Clear() error
	// Read reads the module into memory
	Read() (*storage.Version, error)
}
