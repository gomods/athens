package stash

import (
	"context"
	"fmt"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/gcp"
)

// WithGCSLock returns a distributed singleflight
// using a GCS backend. See the config.toml documentation for details.
func WithGCSLock(staleThreshold int, s storage.Backend) (Wrapper, error) {
	if staleThreshold <= 0 {
		return nil, errors.E("stash.WithGCSLock", fmt.Errorf("invalid stale threshold"))
	}
	// Since we *must* be using a GCP stoagfe backend, we can abuse this
	// fact to mutate it, so that we can get our threshold into Save().
	// Your instincts are correct, this is kind of gross.
	gs, ok := s.(*gcp.Storage)
	if !ok {
		return nil, errors.E("stash.WithGCSLock", fmt.Errorf("GCP singleflight can only be used with GCP storage"))
	}
	gs.SetStaleThreshold(time.Duration(staleThreshold) * time.Second)
	return func(s Stasher) Stasher {
		return &gcsLock{s}
	}, nil
}

type gcsLock struct {
	stasher Stasher
}

func (s *gcsLock) Stash(ctx context.Context, mod, ver string) (newVer string, err error) {
	const op errors.Op = "gcslock.Stash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	newVer, err = s.stasher.Stash(ctx, mod, ver)
	if err != nil {
		// already been saved before, move on.
		if errors.Is(err, errors.KindAlreadyExists) {
			return ver, nil
		}
		return ver, errors.E(op, err)
	}
	return newVer, nil
}
