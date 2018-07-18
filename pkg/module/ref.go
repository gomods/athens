package module

import (
	"github.com/gomods/athens/pkg/storage"
)

// Ref points to a module somewhere
type Ref interface {
	// Clear frees the storage & resources that the module uses
	Clear() error
	// Read reads the module into memory
	Read() (*storage.Version, error)
}
