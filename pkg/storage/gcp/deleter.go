package gcp

import (
	"context"

	"github.com/gomods/athens/pkg/storage"
	m "github.com/gomods/athens/pkg/storage/module"
)

// Delete implements the (./pkg/storage).Deleter interface and
// removes a version of a module from storage. Returning ErrNotFound
// if the version does not exist.
func (s *Storage) Delete(module, version string) error {
	ctx := context.Background()
	if exists := s.Exists(module, version); !exists {
		return storage.ErrVersionNotFound{Module: module, Version: version}
	}

	return m.Delete(ctx, module, version, s.delete)
}

func (s *Storage) delete(ctx context.Context, path string) error {
	return s.bucket.Object(path).Delete(ctx)
}
