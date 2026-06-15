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
		go s.process(ctx, mod, ver)
	} else {
		s.subs[mv] = append(s.subs[mv], subCh)
	}
	s.mu.Unlock()

	resp := <-subCh
	return resp.newVer, resp.err
}
