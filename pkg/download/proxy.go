package download

import (
	"context"

	"github.com/gomods/athens/pkg/goproxy"
	"github.com/gomods/athens/pkg/storage"
)

type proxyLister struct {
	baseURL string

	delegate UpstreamLister
}

func (p *proxyLister) proxyList(ctx context.Context, mod string) ([]string, error) {
	repo, err := goproxy.NewProxyRepo(p.baseURL, mod)
	if err != nil {
		return nil, err
	}
	return repo.Versions("")
}

func (p *proxyLister) proxyLatest(ctx context.Context, mod string) (*storage.RevInfo, error) {
	repo, err := goproxy.NewProxyRepo(p.baseURL, mod)
	if err != nil {
		return nil, err
	}
	return repo.Latest()
}

func (p *proxyLister) List(ctx context.Context, mod string) ([]string, error) {
	list, err := p.proxyList(ctx, mod)
	if err != nil {
		return p.delegate.List(ctx, mod)
	}
	return list, nil
}

func (p *proxyLister) Latest(ctx context.Context, mod string) (*storage.RevInfo, error) {
	latest, err := p.proxyLatest(ctx, mod)
	if err != nil {
		return p.delegate.Latest(ctx, mod)
	}
	return latest, nil
}

func WithProxy(baseURL string, delegate UpstreamLister) UpstreamLister {
	return &proxyLister{
		baseURL:  baseURL,
		delegate: delegate,
	}
}
