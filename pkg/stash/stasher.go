package stash

import (
	"context"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"go.opencensus.io/trace"
)

// Stasher has the job of taking a module
// from an upstream entity and stashing it to a Storage Backend.
// It also returns a string that represents a semver version of
// what was requested, this is helpful if what was requested
// was a descriptive version such as a branch name or a full commit sha.
type Stasher interface {
	Stash(ctx context.Context, mod string, ver string) (string, error)
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
	fetcher module.Fetcher
	storage storage.Backend
}

func (s *stasher) Stash(ctx context.Context, mod, ver string) (string, error) {
	const op errors.Op = "stasher.Stash"
	_, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	// create a new context that ditches whatever deadline the caller passed
	// but keep the tracing info so that we can properly trace the whole thing.
	ctx, cancel := context.WithTimeout(trace.NewContext(context.Background(), span), time.Minute*10)
	defer cancel()

	v, err := s.fetchModule(ctx, mod, ver)
	if err != nil {
		return "", errors.E(op, err)
	}
	defer v.Zip.Close()
	if v.Semver != ver {
		exists, err := s.storage.Exists(ctx, mod, v.Semver)
		if err != nil {
			return "", errors.E(op, err)
		}
		if exists {
			return v.Semver, nil
		}
	}
	err = s.storage.Save(ctx, mod, v.Semver, v.Mod, v.Zip, v.Info)
	if err != nil {
		return "", errors.E(op, err)
	}
	return v.Semver, nil
}

func (s *stasher) fetchModule(ctx context.Context, mod, ver string) (*storage.Version, error) {
	const op errors.Op = "stasher.fetchModule"
	v, err := s.fetcher.Fetch(ctx, mod, ver)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return v, nil
}
