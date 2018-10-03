package download

import (
	"context"
	"io"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/stash"
	"github.com/gomods/athens/pkg/storage"
)

// Protocol is the download protocol which mirrors
// the http requests that cmd/go makes to the proxy.
type Protocol interface {
	// List implements GET /{module}/@v/list
	List(ctx context.Context, mod string) ([]string, error)

	// Info implements GET /{module}/@v/{version}.info
	Info(ctx context.Context, mod, ver string) ([]byte, error)

	// Latest implements GET /{module}/@latest
	Latest(ctx context.Context, mod string) (*storage.RevInfo, error)

	// GoMod implements GET /{module}/@v/{version}.mod
	GoMod(ctx context.Context, mod, ver string) ([]byte, error)

	// Zip implements GET /{module}/@v/{version}.zip
	Zip(ctx context.Context, mod, ver string) (io.ReadCloser, error)
}

// Wrapper helps extend the main protocol's functionality with addons.
type Wrapper func(Protocol) Protocol

// Opts specifies download protocol options to avoid long func signature.
type Opts struct {
	Storage storage.Backend
	Stasher stash.Stasher
	Lister  UpstreamLister
}

// New returns a full implementation of the download.Protocol
// that the proxy needs. New also takes a variadic list of wrappers
// to extend the protocol's functionality (see addons package).
// The wrappers are applied in order, meaning the last wrapper
// passed is the Protocol that gets hit first.
func New(opts *Opts, wrappers ...Wrapper) Protocol {
	var p Protocol = &protocol{opts.Storage, opts.Stasher, opts.Lister}
	for _, w := range wrappers {
		p = w(p)
	}

	return p
}

type protocol struct {
	s       storage.Backend
	stasher stash.Stasher
	lister  UpstreamLister
}

func (p *protocol) List(ctx context.Context, mod string) ([]string, error) {
	const op errors.Op = "protocol.List"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	strList, sErr := p.s.List(ctx, mod)
	// if we got an unexpected storage err then we can not guarantee that the end result contains all versions
	// a tag or repo could have been deleted
	if sErr != nil {
		return nil, errors.E(op, sErr)
	}
	_, goList, goErr := p.lister.List(mod)
	isUnexpGoErr := goErr != nil && !errors.IsRepoNotFoundErr(goErr)
	// if i.e. github is unavailable we should fail as well so that the behavior of the proxy is stable.
	// otherwise we will get different results the next time because i.e. GH is up again
	if isUnexpGoErr {
		return nil, errors.E(op, goErr)
	}

	isRepoNotFoundErr := goErr != nil && errors.IsRepoNotFoundErr(goErr)
	storageEmpty := len(strList) == 0
	if isRepoNotFoundErr && storageEmpty {
		return nil, errors.E(op, errors.M(mod), errors.KindNotFound, goErr)
	}

	return union(goList, strList), nil
}

func (p *protocol) Latest(ctx context.Context, mod string) (*storage.RevInfo, error) {
	const op errors.Op = "protocol.Latest"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	lr, _, err := p.lister.List(mod)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return lr, nil
}

func (p *protocol) Info(ctx context.Context, mod, ver string) ([]byte, error) {
	const op errors.Op = "protocol.Info"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	info, err := p.s.Info(ctx, mod, ver)
	if errors.IsNotFoundErr(err) {
		err = p.stasher.Stash(ctx, mod, ver)
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

func (p *protocol) GoMod(ctx context.Context, mod, ver string) ([]byte, error) {
	const op errors.Op = "protocol.GoMod"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	goMod, err := p.s.GoMod(ctx, mod, ver)
	if errors.IsNotFoundErr(err) {
		err = p.stasher.Stash(ctx, mod, ver)
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

func (p *protocol) Zip(ctx context.Context, mod, ver string) (io.ReadCloser, error) {
	const op errors.Op = "protocol.Zip"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	zip, err := p.s.Zip(ctx, mod, ver)
	if errors.IsNotFoundErr(err) {
		err = p.stasher.Stash(ctx, mod, ver)
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

// union concatenates two version lists and removes duplicates
func union(list1, list2 []string) []string {
	if list1 == nil {
		list1 = []string{}
	}
	if list2 == nil {
		list2 = []string{}
	}
	list := append(list1, list2...)
	unique := []string{}
	m := make(map[string]struct{})
	for _, v := range list {
		if _, ok := m[v]; !ok {
			unique = append(unique, v)
			m[v] = struct{}{}
		}
	}
	return unique
}
