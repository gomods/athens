package storage

import "github.com/gomods/athens/pkg/observ"

// Deleter deletes module metadata and its source from underlying storage
type Deleter interface {
	// Delete must return ErrNotFound if the module/version are not
	// found.
	Delete(ctx observ.ProxyContext, module, vsn string) error
}
