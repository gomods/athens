package azureblob

import (
	"context"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"golang.org/x/sync/errgroup"
)

// Exists implements the (./pkg/storage).Checker interface
// returning true if the module at version exists in storage
func (s *Storage) Exists(ctx context.Context, module string, version string) (bool, error) {
	const op errors.Op = "azureblob.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	fileNames := []string{
		config.PackageVersionedName(module, version, "mod"),  // .mod
		config.PackageVersionedName(module, version, "info"), // .info
		config.PackageVersionedName(module, version, "zip"),  // .zip
	}
	// truth values for the existence of each file
	availabilities := []bool{false, false, false}

	g, ctx := errgroup.WithContext(ctx)
	for i, name := range fileNames {
		g.Go(func() error {
			found, err := s.client.BlobExists(ctx, name)
			if err != nil {
				return err
			}
			availabilities[i] = found
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return false, err
	}

	available := true
	for _, avail := range availabilities {
		available = available && avail
	}
	return available, nil
}
