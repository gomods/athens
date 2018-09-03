package stasher

import (
	"github.com/gomods/athens/pkg/errors"
)

type pooledStasher struct {
	s  Stasher
	ch chan func()
}

// Pooled returns a stasher that runs a stash operation
// {numWorkers} at a time.
func Pooled(fn FactoryFn, numWorkers int) FactoryFn {
	return func(vfn DPVersionFn) Stasher {
		st := &pooledStasher{
			s:  fn(vfn),
			ch: make(chan func()),
		}
		st.start(numWorkers)
		return st
	}
}

func (s *pooledStasher) start(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go s.listen()
	}
}

func (s *pooledStasher) listen() {
	for f := range s.ch {
		f()
	}
}

func (s *pooledStasher) Stash(mod, ver string) error {
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
