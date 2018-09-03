package pool

import (
	"context"
	"io"

	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
)

type protocol struct {
	dp download.Protocol
	ch chan func()
}

// New takes a download Protocol and a number of workers
// and creates a N worker pool that share all the download.Protocol
// methods.
func New(dp download.Protocol, workers int) download.Protocol {
	ch := make(chan func())
	p := &protocol{dp: dp, ch: ch}

	p.start(workers)
	return p
}

func (p *protocol) start(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go p.listen()
	}
}

func (p *protocol) listen() {
	for f := range p.ch {
		f()
	}
}

func (p *protocol) List(ctx context.Context, mod string) ([]string, error) {
	const op errors.Op = "pool.List"
	var vers []string
	var err error
	done := make(chan struct{}, 1)
	p.ch <- func() {
		vers, err = p.dp.List(ctx, mod)
		done <- struct{}{}
	}
	<-done
	if err != nil {
		return nil, errors.E(op, err)
	}

	return vers, nil
}

func (p *protocol) Info(ctx context.Context, mod, ver string) ([]byte, error) {
	const op errors.Op = "pool.Info"
	var info []byte
	var err error
	done := make(chan struct{}, 1)
	p.ch <- func() {
		info, err = p.dp.Info(ctx, mod, ver)
		done <- struct{}{}
	}
	<-done
	if err != nil {
		return nil, errors.E(op, err)
	}
	return info, nil
}

func (p *protocol) Latest(ctx context.Context, mod string) (*storage.RevInfo, error) {
	const op errors.Op = "pool.Latest"
	var info *storage.RevInfo
	var err error
	done := make(chan struct{}, 1)
	p.ch <- func() {
		info, err = p.dp.Latest(ctx, mod)
		done <- struct{}{}
	}
	<-done
	if err != nil {
		return nil, errors.E(op, err)
	}
	return info, nil
}

func (p *protocol) GoMod(ctx context.Context, mod, ver string) ([]byte, error) {
	const op errors.Op = "pool.GoMod"
	var goMod []byte
	var err error
	done := make(chan struct{}, 1)
	p.ch <- func() {
		goMod, err = p.dp.GoMod(ctx, mod, ver)
		done <- struct{}{}
	}
	<-done
	if err != nil {
		return nil, errors.E(op, err)
	}
	return goMod, nil
}

func (p *protocol) Zip(ctx context.Context, mod, ver string) (io.ReadCloser, error) {
	const op errors.Op = "pool.Zip"
	var zip io.ReadCloser
	var err error
	done := make(chan struct{}, 1)
	p.ch <- func() {
		zip, err = p.dp.Zip(ctx, mod, ver)
		done <- struct{}{}
	}
	<-done
	if err != nil {
		return nil, errors.E(op, err)
	}
	return zip, nil
}

func (p *protocol) Version(ctx context.Context, mod, ver string) (*storage.Version, error) {
	const op errors.Op = "pool.Version"
	info, err := p.Info(ctx, mod, ver)
	if err != nil {
		return nil, errors.E(op, err)
	}

	goMod, err := p.GoMod(ctx, mod, ver)
	if err != nil {
		return nil, errors.E(op, err)
	}

	zip, err := p.dp.Zip(ctx, mod, ver)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return &storage.Version{
		Info: info,
		Mod:  goMod,
		Zip:  zip,
	}, nil
}
