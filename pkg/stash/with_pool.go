package stash

import (
	"github.com/gomods/athens/pkg/errors"
)

type withpool struct {
	s Stasher

	// see download/addons/with_pool
	// for design docs on about this channel.
	jobCh chan func()
}

// WithPool returns a stasher that runs a stash operation
// {numWorkers} at a time.
func WithPool(numWorkers int) Wrapper {
	return func(s Stasher) Stasher {
		st := &withpool{
			s:     s,
			jobCh: make(chan func()),
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

func (s *withpool) Stash(mod, ver string) error {
	const op errors.Op = "stash.Pool"
	var err error
	done := make(chan struct{}, 1)
	s.jobCh <- func() {
		err = s.s.Stash(mod, ver)
		close(done)
	}
	<-done
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}
