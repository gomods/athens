package gcp

import (
	"context"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage"
	multierror "github.com/hashicorp/go-multierror"
)

// Delete implements the (./pkg/storage).Deleter interface and
// removes a version of a module from storage. Returning ErrNotFound
// if the version does not exist.
func (s *Storage) Delete(module, version string) error {
	ctx := context.Background()
	if exists := s.bucket.Exists(ctx, config.PackageVersionedName(module, version, "mod")); !exists {
		return storage.ErrVersionNotFound{Module: module, Version: version}
	}
	var errs error
	if err := s.bucket.Delete(ctx, config.PackageVersionedName(module, version, "mod")); err != nil {
		errs = multierror.Append(errs, err)
	}
	if err := s.bucket.Delete(ctx, config.PackageVersionedName(module, version, "info")); err != nil {
		errs = multierror.Append(errs, err)
	}
	if err := s.bucket.Delete(ctx, config.PackageVersionedName(module, version, "zip")); err != nil {
		errs = multierror.Append(errs, err)
	}
	return errs
}
