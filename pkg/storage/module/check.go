package module

import (
	"context"

	"github.com/gomods/athens/pkg/config"
	"golang.org/x/sync/errgroup"
)

type Checker = func(context.Context, string) (bool, error)

func Exists(ctx context.Context, module, version string, check Checker) (bool, error) {
	availableCh := make(chan bool, 3)

	names := []string{
		config.PackageVersionedName(module, version, "mod"),
		config.PackageVersionedName(module, version, "info"),
		config.PackageVersionedName(module, version, "zip"),
	}
	g, ctx := errgroup.WithContext(ctx)
	for _, name := range names {
		// don't remove the below line because you need to close over the name
		// variable inside the goroutine so you don't get a race on the next
		// iteration of the loop
		n := name
		g.Go(func() error {
			found, err := check(ctx, n)
			if err != nil {
				return err
			}
			availableCh <- found
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return false, err
	}

	available := true
	for avail := range availableCh {
		available = available && avail
	}
	return available, nil
}
