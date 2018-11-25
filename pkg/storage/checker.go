package storage

import "github.com/gomods/athens/pkg/observ"

// Checker is the interface that checks if the version of the module exists
type Checker interface {
	// Exists checks whether or not module in specified version is present
	// in the backing storage
	Exists(ctx observ.ProxyContext, module, version string) (bool, error)
}
