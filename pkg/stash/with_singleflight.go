package stash

import (
	"context"
	"sync"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/paths"
)

// WithSingleflight returns a singleflight stasher.
// This two clients make two subsequent
// requests to stash a module, then
// it will only do it once and give the first
// response to both the first and the second client.
func WithSingleflight(s Stasher) Stasher {
	sf := &withsf{}
	sf.s = s
	sf.subs = map[string][]chan error{}

	return sf
}

type withsf struct {
	s Stasher

	mu   sync.Mutex
	subs map[string][]chan error
}

func (s *withsf) process(ctx context.Context, mod, ver string) {
	mv := paths.FmtModVer(mod, ver)
	err := s.s.Stash(ctx, mod, ver)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.subs[mv] {
		ch <- err
	}
	delete(s.subs, mv)
}

func (s *withsf) Stash(ctx context.Context, mod, ver string) error {
	const op errors.Op = "singleflight.Stash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	mv := paths.FmtModVer(mod, ver)
	s.mu.Lock()
	subCh := make(chan error, 1)
	_, inFlight := s.subs[mv]
	if !inFlight {
		s.subs[mv] = []chan error{subCh}
		go s.process(ctx, mod, ver)
	} else {
		s.subs[mv] = append(s.subs[mv], subCh)
	}
	s.mu.Unlock()

	return <-subCh
}
