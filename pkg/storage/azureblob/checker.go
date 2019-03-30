package azureblob

import (
	"context"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// Exists implements the (./pkg/storage).Checker interface
// returning true if the module at version exists in storage
func (s *Storage) Exists(ctx context.Context, module string, version string) (bool, error) {
	const op errors.Op = "azureblob.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	px := config.PackageVersionedName(module, version, "")
	blobs, err := s.client.ListBlobs(ctx, px)
	if err != nil {
		return false, errors.E(op, err, errors.M(module), errors.V(version))
	}

	return len(blobs) == 3, nil
}
