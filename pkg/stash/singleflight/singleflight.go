package singleflight

import (
	"sync"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/stash"
)

// New returns a singleflight stasher.
// This two clients make two subsequent
// requests to stash a module, then
// it will only do it once and give the first
// response to both the first and the second client.
func New(s stash.Stasher) stash.Stasher {
	sf := &stasher{}
	sf.s = s
	sf.mp = map[string]struct{}{}
	sf.subs = map[string][]chan error{}

	return sf
}

type stasher struct {
	s stash.Stasher

	mu   sync.Mutex
	mp   map[string]struct{}
	subs map[string][]chan error
}

func (s *stasher) process(mod, ver string) {
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

func (s *stasher) Stash(mod, ver string) error {
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
