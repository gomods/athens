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
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/stash"
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
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

// Wrapper helps extend the main stasher's functionality with addons.
type Wrapper func(Protocol) Protocol

// Opts specifies download protocol options to avoid long func signature.
type Opts struct {
	Storage   storage.Backend
	Stasher   stash.Stasher
	GoBinPath string
	Fs        afero.Fs
}

// New returns a full implementation of the download.Protocol
// that the proxy needs. New also takes a variadic list of wrappers
// to extend the protocol's functionality (see addons package).
// The wrappers are applied in order, meaning the last wrapper
// passed is the Protocol that gets hit first.
func New(opts *Opts, wrappers ...Wrapper) Protocol {
	var p Protocol = &protocol{opts.Storage, opts.Stasher, opts.GoBinPath, opts.Fs}
	for _, w := range wrappers {
		p = w(p)
	}

	return p
}

type protocol struct {
	s         storage.Backend
	stasher   stash.Stasher
	goBinPath string
	fs        afero.Fs
}

func (p *protocol) List(ctx context.Context, mod string) ([]string, error) {
	const op errors.Op = "download.protocol.List"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	lr, err := p.list(op, mod)
	if err != nil {
		return nil, err
	}

	return lr.Versions, nil
}

func (p *protocol) Latest(ctx context.Context, mod string) (*storage.RevInfo, error) {
	const op errors.Op = "download.protocol.Latest"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
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
	const op errors.Op = "download.protocol.Info"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
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
	const op errors.Op = "download.protocol.GoMod"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
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
	const op errors.Op = "download.protocol.Zip"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
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
