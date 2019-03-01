package download

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
)

// UpstreamLister retrieves a list of available module versions from upstream
// i.e. VCS, and a Storage backend.
type UpstreamLister interface {
	List(ctx context.Context, mod string) ([]string, error)

	Latest(ctx context.Context, mod string) (*storage.RevInfo, error)
}

type listResp struct {
	Path     string
	Version  string
	Versions []string `json:",omitempty"`
	Time     time.Time
}

type vcsLister struct {
	goBinPath string
	fs        afero.Fs
}

func (l *vcsLister) List(ctx context.Context, mod string) ([]string, error) {
	_, versions, err := l.list(ctx, mod)
	return versions, err
}

func (l *vcsLister) Latest(ctx context.Context, mod string) (*storage.RevInfo, error) {
	revInfo, _, err := l.list(ctx, mod)
	return revInfo, err
}

func (l *vcsLister) list(ctx context.Context, mod string) (*storage.RevInfo, []string, error) {
	const op errors.Op = "vcsLister.List"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	hackyPath, err := afero.TempDir(l.fs, "", "hackymod")
	if err != nil {
		return nil, nil, errors.E(op, err)
	}
	defer l.fs.RemoveAll(hackyPath)
	err = module.Dummy(l.fs, hackyPath)
	if err != nil {
		return nil, nil, errors.E(op, err)
	}

	cmd := exec.Command(
		l.goBinPath,
		"list", "-m", "-versions", "-json",
		config.FmtModVer(mod, "latest"),
	)
	cmd.Dir = hackyPath
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	gopath, err := afero.TempDir(l.fs, "", "athens")
	if err != nil {
		return nil, nil, errors.E(op, err)
	}
	defer module.ClearFiles(l.fs, gopath)
	cmd.Env = module.PrepareEnv(gopath)

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("%v: %s", err, stderr)
		return nil, nil, errors.E(op, err)
	}

	var lr listResp
	err = json.NewDecoder(stdout).Decode(&lr)
	if err != nil {
		return nil, nil, errors.E(op, err)
	}
	rev := storage.RevInfo{
		Time:    lr.Time,
		Version: lr.Version,
	}
	return &rev, lr.Versions, nil
}

// NewVCSLister creates an UpstreamLister which uses VCS to fetch a list of available versions
func NewVCSLister(goBinPath string, fs afero.Fs) UpstreamLister {
	return &vcsLister{goBinPath: goBinPath, fs: fs}
}
