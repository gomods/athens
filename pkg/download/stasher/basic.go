package stasher

import (
	"context"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
)

// Stasher has the job of taking a module
// from its upstream Download Protcool and
// stashing to its Storage Backend. Both
// interfaces must be defined on the type previously.
type Stasher interface {
	Stash(mod, ver string) error
}

// DPVersionFn is a version func of download protocol
type DPVersionFn func(context.Context, string, string) (*storage.Version, error)

// FactoryFn is a factory func accepting version func of download protocol resulting in a stasher
type FactoryFn func(DPVersionFn) Stasher

// Basic is a ctor for basic stasher
func Basic(sb storage.Backend) FactoryFn {
	return func(vfn DPVersionFn) Stasher {
		return &stasher{vfn, sb}
	}
}

type stasher struct {
	versionFn DPVersionFn
	s         storage.Backend
}

func (s *stasher) Stash(mod, ver string) error {
	const op errors.Op = "stasher.Stash"
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	if e, _ := s.s.Exists(ctx, mod, ver); e == true {
		return nil
	}

	v, err := s.versionFn(ctx, mod, ver)
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
