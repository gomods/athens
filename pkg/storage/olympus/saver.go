package olympus

import (
	"context"

	"github.com/gomods/athens/pkg/storage"
)

// asserts that a ModuleStore is a storage.Saver
var _ storage.Saver = &ModuleStore{}

// Save stores a module in olympus.
// This actually does not store anything just reports cache miss
func (s *ModuleStore) Save(_ context.Context, module, version string, _, _, _ []byte) error {
	// dummy implementation so Olympus Store can be used everywhere as Backend iface
	return nil
}
