package stash

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// WithGCSLock returns a distributed singleflight
// using a GCS backend. See the config.toml documentation for details.
func WithGCSLock(s Stasher) Stasher {
	return &gcsLock{s}
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
