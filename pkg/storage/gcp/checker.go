package gcp

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"google.golang.org/api/iterator"
)

// Exists implements the (./pkg/storage).Checker interface
// returning true if the module at version exists in storage
func (s *Storage) Exists(ctx context.Context, module, version string) (bool, error) {
	const op errors.Op = "gcp.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	it := s.bucket.Objects(ctx, &storage.Query{Prefix: config.PackageVersionedName(module, version, "")})
	var count int
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return false, errors.E(op, err, errors.M(module), errors.V(version))
		}
		switch attrs.Name {
		case config.PackageVersionedName(module, version, "info"):
			count++
		case config.PackageVersionedName(module, version, "mod"):
			count++
		case config.PackageVersionedName(module, version, "zip"):
			count++
		}
	}

	return count == 3, nil
}
