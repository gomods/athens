package stash

import (
	"context"
	"sync"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// WithSingleflight returns a singleflight stasher.
// This two clients make two subsequent
// requests to stash a module, then
// it will only do it once and give the first
// response to both the first and the second client.
func WithSingleflight(s Stasher) Stasher {
	sf := &withsf{}
	sf.stasher = s
	sf.subs = map[string][]chan *sfResp{}

	return sf
}

// WithLocalSingleflight returns a Wrapper that applies the distributed wrapper
// and then coalesces, in this process, concurrent calls for the same
// module@version in front of it. They share one call into distributed, so they
// take its lock once and share its result and its error, instead of queueing up
// for the lock one after another and each retrying a failed stash in turn.
//
// The flight is detached from the callers' cancellation: it keeps going, up to
// the first caller's deadline, when every caller has gone. That includes
// waiting for the distributed lock, which a lone caller would previously have
// given up on as soon as it disconnected.
func WithLocalSingleflight(distributed Wrapper) Wrapper {
	return func(s Stasher) Stasher {
		return WithSingleflight(distributed(s))
	}
}

type sfResp struct {
	newVer string
	err    error
}

type withsf struct {
	stasher Stasher

	mu   sync.Mutex
	subs map[string][]chan *sfResp
}

func (s *withsf) process(ctx context.Context, mod, ver string) {
	mv := config.FmtModVer(mod, ver)
	newVer, err := s.stasher.Stash(ctx, mod, ver)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.subs[mv] {
		ch <- &sfResp{newVer, err}
	}
	delete(s.subs, mv)
}

func (s *withsf) Stash(ctx context.Context, mod, ver string) (string, error) {
	const op errors.Op = "singleflight.Stash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	mv := config.FmtModVer(mod, ver)
	s.mu.Lock()
	subCh := make(chan *sfResp, 1)
	_, inFlight := s.subs[mv]
	if !inFlight {
		s.subs[mv] = []chan *sfResp{subCh}
		// The work is shared by every subscriber, so it must not be
		// cancelled just because the waiter that started it went away.
		// WithoutCancel also drops the deadline, so carry it over: the flight
		// stays bounded and a hung stash cannot wedge this module@version.
		workCtx, cancelWork := context.WithoutCancel(ctx), context.CancelFunc(func() {})
		if deadline, ok := ctx.Deadline(); ok {
			workCtx, cancelWork = context.WithDeadline(workCtx, deadline)
		}
		go func() {
			defer cancelWork()
			s.process(workCtx, mod, ver)
		}()
	} else {
		s.subs[mv] = append(s.subs[mv], subCh)
	}
	s.mu.Unlock()

	// subCh is buffered, so process never blocks on a waiter that has left.
	select {
	case resp := <-subCh:
		return resp.newVer, resp.err
	case <-ctx.Done():
		return "", errors.E(op, ctx.Err())
	}
}
