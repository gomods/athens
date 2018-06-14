package olympus

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/storage"
)

// asserts that a ModuleStore is a storage.Saver
var _ storage.Saver = &ModuleStore{}

// Save stores a module in olympus.
// This actually does not store anything just reports cache miss
func (s *ModuleStore) Save(_ buffalo.Context, module, version string, _, _, _ []byte) error {
	// dummy implementation so Olympus Store can be used everywhere as Backend iface
	return nil
}
