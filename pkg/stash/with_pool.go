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
	for i := 0; i < numWorkers; i++ {
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
	var err error
	var newVer string
	done := make(chan struct{}, 1)
	s.jobCh <- func() {
		newVer, err = s.stasher.Stash(ctx, mod, ver)
		close(done)
	}
	<-done
	if err != nil {
		return "", errors.E(op, err)
	}

	return newVer, nil
}
