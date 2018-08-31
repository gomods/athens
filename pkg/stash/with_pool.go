package stash

import (
	"github.com/gomods/athens/pkg/errors"
)

type withpool struct {
	s  Stasher
	ch chan func()
}

// WithPool returns a stasher that runs a stash operation
// {numWorkers} at a time.
func WithPool(numWorkers int) Wrapper {
	return func(s Stasher) Stasher {
		st := &withpool{
			s:  s,
			ch: make(chan func()),
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
	for f := range s.ch {
		f()
	}
}

func (s *withpool) Stash(mod, ver string) error {
	const op errors.Op = "stash.Pool"
	var err error
	done := make(chan struct{}, 1)
	s.ch <- func() {
		err = s.s.Stash(mod, ver)
		done <- struct{}{}
	}
	<-done
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}
