package module

import (
	"context"
	"encoding/json"

	"github.com/gomods/athens/pkg/goproxy"
	"github.com/gomods/athens/pkg/storage"
)

type proxyFetcher struct {
	baseURL string

	delegate Fetcher
}

func (p *proxyFetcher) proxyFetch(ctx context.Context, mod, ver string) (*storage.Version, error) {
	proxyRepo, err := goproxy.NewProxyRepo(p.baseURL, mod)
	if err != nil {
		return nil, err
	}
	modBytes, err := proxyRepo.GoMod(ver)
	if err != nil {
		return nil, err
	}
	zipReader, err := proxyRepo.Zip(ver)
	if err != nil {
		return nil, err
	}
	infoBytes, err := proxyRepo.Stat(ver)
	if err != nil {
		return nil, err
	}
	info := new(storage.RevInfo)
	if err := json.Unmarshal(infoBytes, info); err != nil {
		return nil, err
	}
	return &storage.Version{
		Mod:    modBytes,
		Zip:    zipReader,
		Info:   infoBytes,
		Semver: info.Version,
	}, nil
}

func (p *proxyFetcher) delegateFetch(ctx context.Context, mod, ver string) (*storage.Version, error) {
	return p.delegate.Fetch(ctx, mod, ver)
}

func (p *proxyFetcher) Fetch(ctx context.Context, mod, ver string) (*storage.Version, error) {
	version, e := p.proxyFetch(ctx, mod, ver)
	if e != nil {
		return p.delegateFetch(ctx, mod, ver)
	}
	return version, nil
}

func WithProxy(baseURL string, delegate Fetcher) Fetcher {
	return &proxyFetcher{
		baseURL:  baseURL,
		delegate: delegate,
	}
}
