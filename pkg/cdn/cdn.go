package cdn

import (
	"fmt"
)

// ErrNotFound ios an error implementation to indicate that a module is not
// found
type ErrNotFound struct {
	ModuleName    string
	ModuleVersion string
}

// Error is the error interface implementation
func (e ErrNotFound) Error() string {
	return fmt.Sprintf("Module %s/%s not found", e.ModuleName, e.ModuleVersion)
}

// CDN represents access to a CDN for registry use
type CDN interface {
	Saver
	Getter
}
