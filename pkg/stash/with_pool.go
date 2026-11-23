package stash

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

type withpool struct {
	stasher Stasher

	// see download/addons/with_pool
	// for design docs on about this channel.
	jobCh chan func()
}

// WithPool returns a stasher that runs a stash operation
// {numWorkers} at a time.
func WithPool(numWorkers int) Wrapper {
	return func(s Stasher) Stasher {
		st := &withpool{
			stasher: s,
			jobCh:   make(chan func()),
		}
		st.start(numWorkers)
		return st
	}
}

func (s *withpool) start(numWorkers int) {
	for range numWorkers {
		go s.listen()
	}
}

func (s *withpool) listen() {
	for f := range s.jobCh {
		f()
	}
}

func (s *withpool) Stash(ctx context.Context, mod, ver string) (string, error) {
	const op errors.Op = "stash.Pool"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	// The result is passed through a buffered channel rather than captured
	// variables so that the job can finish after Stash has returned early
	// without racing with the caller and without blocking the worker.
	resCh := make(chan poolResp, 1)
	job := func() {
		newVer, err := s.stasher.Stash(ctx, mod, ver)
		resCh <- poolResp{newVer: newVer, err: err}
	}

	select {
	case s.jobCh <- job:
	case <-ctx.Done():
		return "", errors.E(op, ctx.Err())
	}

	select {
	case res := <-resCh:
		if res.err != nil {
			return "", errors.E(op, res.err)
		}
		return res.newVer, nil
	case <-ctx.Done():
		return "", errors.E(op, ctx.Err())
	}
}

type poolResp struct {
	newVer string
	err    error
}
