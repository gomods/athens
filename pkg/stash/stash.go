package stash

import (
	"context"
	"time"

	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
)

// Stasher has the job of taking a module
// from its upstream Sownload Protcool and
// stashing to its Storage Backend. Both
// interfaces must be defined on the type previously.
type Stasher interface {
	Stash(mod, ver string) error
}

// New returns a plain stasher that takes
// a module from a download.Protocol and
// stashes it into a backend.Storage.
func New(dp download.Protocol, s storage.Backend) Stasher {
	return &stasher{dp, s}
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
