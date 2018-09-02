package download

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/stash"
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
)

// New returns a full implementation of the download.Protocol
// that the proxy needs.
func New(s storage.Backend, st stash.Stasher, goBinPath string, fs afero.Fs) Protocol {
	return &protocol{s, st, goBinPath, fs}
}

type protocol struct {
	s         storage.Backend
	stasher   stash.Stasher
	goBinPath string
	fs        afero.Fs
}

func (p *protocol) List(ctx context.Context, mod string) ([]string, error) {
	const op errors.Op = "goget.List"
	lr, err := p.list(op, mod)
	if err != nil {
		return nil, err
	}

	return lr.Versions, nil
}

func (p *protocol) Latest(ctx context.Context, mod string) (*storage.RevInfo, error) {
	const op errors.Op = "goget.Latest"
	lr, err := p.list(op, mod)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return &storage.RevInfo{
		Time:    lr.Time,
		Version: lr.Version,
	}, nil
}

type listResp struct {
	Path     string
	Version  string
	Versions []string `json:",omitempty"`
	Time     time.Time
}

func (p *protocol) list(op errors.Op, mod string) (*listResp, error) {
	hackyPath, err := afero.TempDir(p.fs, "", "hackymod")
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer p.fs.RemoveAll(hackyPath)
	err = module.Dummy(p.fs, hackyPath)
	if err != nil {
		return nil, errors.E(op, err)
	}

	cmd := exec.Command(
		p.goBinPath,
		"list", "-m", "-versions", "-json",
		config.FmtModVer(mod, "latest"),
	)
	cmd.Dir = hackyPath
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	gopath, err := afero.TempDir(p.fs, "", "athens")
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer module.ClearFiles(p.fs, gopath)
	cmd.Env = module.PrepareEnv(gopath)

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("%v: %s", err, stderr)
		return nil, errors.E(op, err)
	}

	var lr listResp
	err = json.NewDecoder(stdout).Decode(&lr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return &lr, nil
}

func (p *protocol) Info(ctx context.Context, mod, ver string) ([]byte, error) {
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

func (p *protocol) GoMod(ctx context.Context, mod, ver string) ([]byte, error) {
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

func (p *protocol) Zip(ctx context.Context, mod, ver string) (io.ReadCloser, error) {
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
