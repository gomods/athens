package cdn

import (
	"github.com/gomods/athens/pkg/eventlog"
	"github.com/gomods/athens/pkg/storage"
)

// Saver saves a module metadata & storage to its underlying storage
type Saver interface {
	// Save saves the given module and version to the CDN
	Save(module string, version *storage.Version) error
	// SaveMetadata saves "op" to the module key/value store for module/version
	SaveMetadata(module, version string, op eventlog.Op) error
}
