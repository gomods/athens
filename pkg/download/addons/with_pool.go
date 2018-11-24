package addons

import (
	"context"
	"io"

	"github.com/gomods/athens/pkg/paths"

	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
)

type withpool struct {
	dp download.Protocol

	// jobCh is a channel that takes an anonymous
	// function that it executes based on the pool's
	// business. The design levarages closures
	// so that the worker does not need to worry about
	// what the type of job it is taking (Info, Zip etc),
	// it just regulates functions and executes them
	// in a worker-pool fashion.
	jobCh chan func()
}

// WithPool takes a download Protocol and a number of workers
// and creates a N worker pool that share all the download.Protocol
// methods.
func WithPool(workers int) download.Wrapper {
	return func(dp download.Protocol) download.Protocol {
		jobCh := make(chan func())
		p := &withpool{dp: dp, jobCh: jobCh}

		p.start(workers)
		return p
	}
}

func (p *withpool) start(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go p.listen()
	}
}

func (p *withpool) listen() {
	for f := range p.jobCh {
		f()
	}
}

func (p *withpool) List(ctx context.Context, mod string) ([]string, error) {
	const op errors.Op = "pool.List"
	var vers []string
	var err error
	done := make(chan struct{}, 1)
	p.jobCh <- func() {
		vers, err = p.dp.List(ctx, mod)
		close(done)
	}
	<-done
	if err != nil {
		return nil, errors.E(op, err)
	}

	return vers, nil
}

func (p *withpool) Info(ctx context.Context, mod, ver string) ([]byte, error) {
	const op errors.Op = "pool.Info"
	var info []byte
	var err error
	done := make(chan struct{}, 1)
	p.jobCh <- func() {
		info, err = p.dp.Info(ctx, mod, ver)
		close(done)
	}
	<-done
	if err != nil {
		return nil, errors.E(op, err)
	}
	return info, nil
}

func (p *withpool) Latest(ctx context.Context, mod string) (*storage.RevInfo, error) {
	const op errors.Op = "pool.Latest"
	var info *storage.RevInfo
	var err error
	done := make(chan struct{}, 1)
	p.jobCh <- func() {
		info, err = p.dp.Latest(ctx, mod)
		close(done)
	}
	<-done
	if err != nil {
		return nil, errors.E(op, err)
	}
	return info, nil
}

func (p *withpool) GoMod(ctx context.Context, mod, ver string) ([]byte, error) {
	const op errors.Op = "pool.GoMod"
	var goMod []byte
	var err error
	done := make(chan struct{}, 1)
	p.jobCh <- func() {
		goMod, err = p.dp.GoMod(ctx, mod, ver)
		close(done)
	}
	<-done
	if err != nil {
		return nil, errors.E(op, err)
	}
	return goMod, nil
}

func (p *withpool) Zip(ctx context.Context, mod, ver string) (io.ReadCloser, error) {
	const op errors.Op = "pool.Zip"
	var zip io.ReadCloser
	var err error
	done := make(chan struct{}, 1)
	p.jobCh <- func() {
		zip, err = p.dp.Zip(ctx, mod, ver)
		close(done)
	}
	<-done
	if err != nil {
		return nil, errors.E(op, err)
	}
	return zip, nil
}

func (p *withpool) Catalog(ctx context.Context, token string, numElement int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "pool.Catalog"
	var modsVers []paths.AllPathParams
	var nextToken string
	var err error
	done := make(chan struct{}, 1)
	p.jobCh <- func() {
		modsVers, nextToken, err = p.dp.Catalog(ctx, token, numElement)
		close(done)
	}
	<-done
	if err != nil {
		return nil, "", errors.E(op, err)
	}

	return modsVers, nextToken, nil
}
