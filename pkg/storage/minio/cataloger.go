package minio

import (
	"context"

	"github.com/gomods/athens/pkg/paths"

	"github.com/gomods/athens/pkg/errors"
)

// Catalog implements the (./pkg/storage).Cataloger interface
// It returns a list of modules and versions contained in the storage
func (s *storageImpl) Catalog(ctx context.Context, token string, elements int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "minio.Catalog"
	return nil, "", errors.E(op, errors.KindMethodNotImplemented)
}
