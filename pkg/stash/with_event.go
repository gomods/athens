package stash

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/events"
)

// WithEventsHook returns a stasher that can send out Stashed events
// to the given implementation
func WithEventsHook(e events.Hook) Wrapper {
	return func(s Stasher) Stasher {
		return &withEvent{s, e}
	}
}

type withEvent struct {
	s Stasher
	e events.Hook
}

func (we *withEvent) Stash(ctx context.Context, mod string, ver string) (string, error) {
	const op errors.Op = "stash.withEvent"
	resolvedVer, err := we.s.Stash(ctx, mod, ver)
	if err != nil {
		return "", errors.E(op, err)
	}
	err = we.e.Stashed(ctx, mod, resolvedVer)
	if err != nil {
		return "", errors.E(op, err)
	}
	return resolvedVer, nil
}
