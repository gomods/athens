package azureblob

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	modupl "github.com/gomods/athens/pkg/storage/module"
)

// Delete implements the (./pkg/storage).Deleter interface and
// removes a version of a module from storage. Returning ErrNotFound
// if the version does not exist.
func (s *Storage) Delete(ctx context.Context, module string, version string) error {
	const op errors.Op = "azureblob.Delete"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	_, err := s.Info(ctx, module, version)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	return modupl.Delete(ctx, module, version, s.client.DeleteBlob, s.timeout)
}
