package gcp

import (
	"context"
	"fmt"
)

// Exists implements the (./pkg/storage).Checker interface
// returning true if the module at version exists in storage
func (s *Storage) Exists(module, version string) bool {
	ctx := context.Background()
	modName := fmt.Sprintf("%s/@v/%s.%s", module, version, "mod")
	modHandle := s.bucket.Object(modName)
	_, err := modHandle.Attrs(ctx)
	// Unless the signature changes for Exists just say false on any error.
	// Attrs will error with not found if it doesn't exist.
	if err != nil {
		return false
	}
	return true
}
