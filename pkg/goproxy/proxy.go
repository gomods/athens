// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// see github.com/golang/go/src/cmd/go/internal/modfetch/proxy.go

package goproxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/gomods/athens/pkg/storage"
	"github.com/rogpeppe/go-internal/module"
	"github.com/rogpeppe/go-internal/semver"
)

type ProxyRepo struct {
	url string
}

func NewProxyRepo(baseURL, path string) (*ProxyRepo, error) {
	enc, err := module.EncodePath(path)
	if err != nil {
		return nil, err
	}
	return &ProxyRepo{strings.TrimSuffix(baseURL, "/") + "/" + pathEscape(enc)}, nil
}

func (p *ProxyRepo) Versions(prefix string) ([]string, error) {
	var data []byte
	err := webGetBytes(p.url+"/@v/list", &data)
	if err != nil {
		return nil, err
	}
	var list []string
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) >= 1 && semver.IsValid(f[0]) && strings.HasPrefix(f[0], prefix) {
			list = append(list, f[0])
		}
	}
	return list, nil
}

func (p *ProxyRepo) latest() (*storage.RevInfo, error) {
	var data []byte
	err := webGetBytes(p.url+"/@v/list", &data)
	if err != nil {
		return nil, err
	}
	var best time.Time
	var bestVersion string
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) >= 2 && semver.IsValid(f[0]) {
			ft, err := time.Parse(time.RFC3339, f[1])
			if err == nil && best.Before(ft) {
				best = ft
				bestVersion = f[0]
			}
		}
	}
	if bestVersion == "" {
		return nil, fmt.Errorf("no commits")
	}
	info := &storage.RevInfo{
		Version: bestVersion,
		Time:    best,
	}
	return info, nil
}

func (p *ProxyRepo) Stat(rev string) ([]byte, error) {
	var data []byte
	encRev, err := module.EncodeVersion(rev)
	if err != nil {
		return nil, err
	}
	err = webGetBytes(p.url+"/@v/"+pathEscape(encRev)+".info", &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (p *ProxyRepo) Latest() (*storage.RevInfo, error) {
	var data []byte
	u := p.url + "/@latest"
	err := webGetBytes(u, &data)
	if err != nil {
		// TODO return err if not 404
		return p.latest()
	}
	info := new(storage.RevInfo)
	if err := json.Unmarshal(data, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (p *ProxyRepo) GoMod(version string) ([]byte, error) {
	var data []byte
	encVer, err := module.EncodeVersion(version)
	if err != nil {
		return nil, err
	}
	err = webGetBytes(p.url+"/@v/"+pathEscape(encVer)+".mod", &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (p *ProxyRepo) Zip(version string) (io.ReadCloser, error) {
	var body io.ReadCloser
	encVer, err := module.EncodeVersion(version)
	if err != nil {
		return nil, err
	}
	err = webGetBody(p.url+"/@v/"+pathEscape(encVer)+".zip", &body)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	return body, nil
}

// pathEscape escapes s so it can be used in a path.
// That is, it escapes things like ? and # (which really shouldn't appear anyway).
// It does not escape / to %2F: our REST API is designed so that / can be left as is.
func pathEscape(s string) string {
	return strings.ReplaceAll(url.PathEscape(s), "%2F", "/")
}
