package pool

import (
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/stash"
)

type stasher struct {
	s  stash.Stasher
	ch chan func()
}

// New returns a stasher that runs a stash operation
// {numWorkers} at a time.
func New(s stash.Stasher, numWorkers int) stash.Stasher {
	st := &stasher{
		s:  s,
		ch: make(chan func()),
	}
	st.start(numWorkers)
	return st
}

func (s *stasher) start(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go s.listen()
	}
}

func (s *stasher) listen() {
	for f := range s.ch {
		f()
	}
}

func (s *stasher) Stash(mod, ver string) error {
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
