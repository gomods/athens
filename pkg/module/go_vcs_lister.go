package module

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
	"golang.org/x/sync/singleflight"
)

type listResp struct {
	Path     string
	Version  string
	Versions []string `json:",omitempty"`
	Time     time.Time
}

type vcsLister struct {
	goBinPath string
	env       []string
	fs        afero.Fs
	sfg       *singleflight.Group
	timeout   time.Duration
}

// NewVCSLister creates an UpstreamLister which uses VCS to fetch a list of available versions.
func NewVCSLister(goBinPath string, env []string, fs afero.Fs, timeout time.Duration) UpstreamLister {
	return &vcsLister{
		goBinPath: goBinPath,
		env:       env,
		fs:        fs,
		sfg:       &singleflight.Group{},
		timeout:   timeout,
	}
}

type listSFResp struct {
	rev      *storage.RevInfo
	versions []string
}

func (l *vcsLister) List(ctx context.Context, module string) (*storage.RevInfo, []string, error) {
	const op errors.Op = "vcsLister.List"
	_, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	sfResp, err, _ := l.sfg.Do(module, func() (any, error) {
		tmpDir, err := afero.TempDir(l.fs, "", "go-list")
		if err != nil {
			return nil, errors.E(op, err)
		}
		defer func() { _ = l.fs.RemoveAll(tmpDir) }()

		timeoutCtx, cancel := context.WithTimeout(ctx, l.timeout)
		defer cancel()

		cmd := exec.CommandContext(
			timeoutCtx,
			l.goBinPath,
			"list", "-m", "-versions", "-json",
			config.FmtModVer(module, "latest"),
		)
		cmd.Dir = tmpDir
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		gopath, err := afero.TempDir(l.fs, "", "athens")
		if err != nil {
			return nil, errors.E(op, err)
		}
		defer func() { _ = clearFiles(l.fs, gopath) }()
		cmd.Env = prepareEnv(gopath, l.env)

		err = cmd.Run()
		if err != nil && !errors.IsNoChildProcessesErr(err) {
			err = fmt.Errorf("%w: %s", err, stderr)
			if errors.IsErr(timeoutCtx.Err(), context.DeadlineExceeded) {
				return nil, errors.E(op, err, errors.KindGatewayTimeout)
			}

			// as of now, we can't recognize between a true NotFound
			// and an unexpected error, so we choose the more
			// hopeful path of NotFound. This way the Go command
			// will not log en error and we still get to log
			// what happened here if someone wants to dig in more.
			// Once, https://github.com/golang/go/issues/30134 is
			// resolved, we can hopefully differentiate.
			return nil, errors.E(op, err, errors.KindNotFound)
		}

		var lr listResp
		err = json.NewDecoder(stdout).Decode(&lr)
		if err != nil {
			return nil, errors.E(op, err)
		}
		rev := storage.RevInfo{
			Time:    lr.Time,
			Version: lr.Version,
		}
		return listSFResp{
			rev:      &rev,
			versions: lr.Versions,
		}, nil
	})
	if err != nil {
		return nil, nil, err
	}
	ret := sfResp.(listSFResp)
	return ret.rev, ret.versions, nil
}
