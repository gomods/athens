package azureblob

import (
	"context"
	"fmt"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// List implements the (./pkg/storage).Lister interface
// It returns a list of versions, if any, for a given module
func (s *Storage) List(ctx context.Context, module string) ([]string, error) {
	const op errors.Op = "azureblob.List"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	return []string{}, errors.E(op, fmt.Errorf("Not Implemented"), errors.M(module))
}
