package module

import (
	"github.com/gomods/athens/pkg/storage"
)

// Ref points to a module somewhere
type Ref interface {
	// Clear frees the storage & resources that the module uses. Calls to Read after you call
	// this function may fail, regardless of whether this function returns nil or not
	Clear() error
	// Read reads the module into memory and returns it. The caller should call
	// the returned function.
	Read() (*storage.Version, error)
}
