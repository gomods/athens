package azureblob

import (
	"context"
	"fmt"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// Delete implements the (./pkg/storage).Deleter interface and
// removes a version of a module from storage. Returning ErrNotFound
// if the version does not exist.
func (s *Storage) Delete(ctx context.Context, module string, vsn string) error {
	const op errors.Op = "azureblob.newBlobStoreClient.Delete"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	return errors.E(op, fmt.Errorf("Not Implemented"), errors.M(module))
}
