package stash

import (
	"context"
	"time"

	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
)

// Stasher does it
type Stasher interface {
	Stash(string, string) error
}

// Wrapper helps extend the main stasher's functionality with addons.
type Wrapper func(Stasher) Stasher

// New returns a plain stasher that takes
// a module from a download.Protocol and
// stashes it into a backend.Storage.
func New(dp download.Protocol, s storage.Backend, wrappers ...Wrapper) Stasher {
	var st Stasher = &stasher{dp, s}
	for _, w := range wrappers {
		st = w(st)
	}

	return st
}

type stasher struct {
	dp download.Protocol
	s  storage.Backend
}

func (s *stasher) Stash(mod, ver string) error {
	const op errors.Op = "stasher.Stash"
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	v, err := s.dp.Version(ctx, mod, ver)
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
