package stash

import (
	"context"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/storage"
)

// Stasher has the job of taking a module
// from an upstream entity and stashing it to a Storage Backend.
type Stasher interface {
	Stash(string, string) error
}

// Wrapper helps extend the main stasher's functionality with addons.
type Wrapper func(Stasher) Stasher

// New returns a plain stasher that takes
// a module from a download.Protocol and
// stashes it into a backend.Storage.
func New(f module.Fetcher, s storage.Backend, wrappers ...Wrapper) Stasher {
	var st Stasher = &stasher{f, s}
	for _, w := range wrappers {
		st = w(st)
	}

	return st
}

type stasher struct {
	f module.Fetcher
	s storage.Backend
}

func (s *stasher) Stash(mod, ver string) error {
	const op errors.Op = "stasher.Stash"
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	v, err := s.fetchModule(ctx, mod, ver)
	if err != nil {
		return errors.E(op, err)
	}
	defer v.Zip.Close()
	err = s.s.Save(ctx, mod, ver, v.Mod, v.Zip, v.Info)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (s *stasher) fetchModule(ctx context.Context, mod, ver string) (*storage.Version, error) {
	const op errors.Op = "stasher.fetchModule"
	v, err := s.f.Fetch(ctx, mod, ver)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return v, nil
}
