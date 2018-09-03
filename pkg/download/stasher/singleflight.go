package stasher

import (
	"sync"

	"github.com/gomods/athens/pkg/config"
)

// Deduplicated returns a singleflight stasher.
// This two clients make two subsequent
// requests to stash a module, then
// it will only do it once and give the first
// response to both the first and the second client.
func Deduplicated(fn FactoryFn) FactoryFn {
	return func(vfn DPVersionFn) Stasher {
		sf := &dedupStasher{}
		sf.s = fn(vfn)
		sf.mp = map[string]struct{}{}
		sf.subs = map[string][]chan error{}

		return sf
	}
}

type dedupStasher struct {
	s Stasher

	mu   sync.Mutex
	mp   map[string]struct{}
	subs map[string][]chan error
}

func (s *dedupStasher) Stash(mod, ver string) error {
	mv := config.FmtModVer(mod, ver)
	s.mu.Lock()
	subCh := make(chan error, 1)
	s.subs[mv] = append(s.subs[mv], subCh)
	_, ok := s.mp[mv]
	if !ok {
		s.mp[mv] = struct{}{}
		go s.process(mod, ver)
	}
	s.mu.Unlock()

	return <-subCh
}

func (s *dedupStasher) process(mod, ver string) {
	mv := config.FmtModVer(mod, ver)
	err := s.s.Stash(mod, ver)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.subs[mv] {
		ch <- err
	}
	delete(s.subs, mv)
	delete(s.mp, mv)
}
