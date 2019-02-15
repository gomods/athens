package module

import (
	"context"

	"github.com/gomods/athens/pkg/config"
	"golang.org/x/sync/errgroup"
)

type Checker = func(context.Context, string) (bool, error)

func Exists(ctx context.Context, module, version string, check Checker) (bool, error) {
	availabilities := make([]bool, 3)

	names := []string{
		config.PackageVersionedName(module, version, "mod"),
		config.PackageVersionedName(module, version, "info"),
		config.PackageVersionedName(module, version, "zip"),
	}
	g, ctx := errgroup.WithContext(ctx)
	for i, name := range names {
		g.Go(func() error {
			found, err := check(ctx, name)
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
