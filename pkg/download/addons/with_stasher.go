package addons

import (
	"context"
	"io"

	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/stash"
	"github.com/gomods/athens/pkg/storage"
)

type withstasher struct {
	s       storage.Backend
	dp      download.Protocol
	stasher stash.Stasher
}

// WithStasher takes an upstream Protocol and storage
// it always prefers storage, otherwise it goes to upstream
// and stashes the storage with the results through the given stasher.
func WithStasher(dp download.Protocol, s storage.Backend, stasher stash.Stasher) download.Protocol {
	p := &withstasher{dp: dp, s: s, stasher: stasher}

	return p
}

func (p *withstasher) List(ctx context.Context, mod string) ([]string, error) {
	return p.dp.List(ctx, mod)
}

func (p *withstasher) Info(ctx context.Context, mod, ver string) ([]byte, error) {
	const op errors.Op = "stasher.Info"
	info, err := p.s.Info(ctx, mod, ver)
	if errors.IsNotFoundErr(err) {
		err = p.stasher.Stash(mod, ver)
		if err != nil {
			return nil, errors.E(op, err)
		}
		info, err = p.s.Info(ctx, mod, ver)
	}
	if err != nil {
		return nil, errors.E(op, err)
	}

	return info, nil
}

func (p *withstasher) Latest(ctx context.Context, mod string) (*storage.RevInfo, error) {
	const op errors.Op = "stasher.Latest"
	info, err := p.dp.Latest(ctx, mod)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return info, nil
}

func (p *withstasher) GoMod(ctx context.Context, mod, ver string) ([]byte, error) {
	const op errors.Op = "stasher.GoMod"
	goMod, err := p.s.GoMod(ctx, mod, ver)
	if errors.IsNotFoundErr(err) {
		err = p.stasher.Stash(mod, ver)
		if err != nil {
			return nil, errors.E(op, err)
		}
		goMod, err = p.s.GoMod(ctx, mod, ver)
	}
	if err != nil {
		return nil, errors.E(op, err)
	}

	return goMod, nil
}

func (p *withstasher) Zip(ctx context.Context, mod, ver string) (io.ReadCloser, error) {
	const op errors.Op = "stasher.Zip"
	zip, err := p.s.Zip(ctx, mod, ver)
	if errors.IsNotFoundErr(err) {
		err = p.stasher.Stash(mod, ver)
		if err != nil {
			return nil, errors.E(op, err)
		}
		zip, err = p.s.Zip(ctx, mod, ver)
	}
	if err != nil {
		return nil, errors.E(op, err)
	}

	return zip, nil
}
