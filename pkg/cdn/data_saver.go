package cdn

import (
	"github.com/gomods/athens/pkg/storage"
)

// DataSaver saves a module metadata & storage to its underlying storage
type DataSaver interface {
	// Save saves the given module and metadata and source code to
	// the CDN and records its location in module metadata storage
	Save(module string, version *storage.Version) error
}
