package gcp

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// Exists implements the (./pkg/storage).Checker interface
// returning true if the module at version exists in storage
func (s *Storage) Exists(ctx context.Context, module, version string) (bool, error) {
	const op errors.Op = "gcp.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	_, err := s.bucket.Object(config.PackageVersionedName(module, version, "mod")).Attrs(ctx)

	if err == storage.ErrObjectNotExist {
		return false, nil
	}

	if err != nil {
		return false, errors.E(op, err, errors.M(module), errors.V(version))
	}

	return true, nil
}
