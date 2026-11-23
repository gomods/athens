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

		runErr := cmd.Run()

		// `go list -m -versions -json` writes the version list as a JSON object
		// on stdout on success, and writes diagnostics to stderr on failure. We
		// decide success/failure from whether stdout parsed, treating the exit
		// code as advisory only.
		var lr listResp
		jsonErr := json.NewDecoder(stdout).Decode(&lr)
		if jsonErr == nil {
			rev := storage.RevInfo{
				Time:    lr.Time,
				Version: lr.Version,
			}
			return listSFResp{
				rev:      &rev,
				versions: lr.Versions,
			}, nil
		}

		// `go list` did not produce a parseable version list. Build a
		// diagnostic from stderr (populated for real failures) and fall back to
		// jsonErr (e.g. io.EOF) when stderr is empty.
		listErr := jsonErr
		if runErr != nil {
			listErr = fmt.Errorf("%w: %s", runErr, stderr)
		} else if stderr.Len() > 0 {
			listErr = fmt.Errorf("%w: %s", jsonErr, stderr)
		}

		if errors.IsErr(timeoutCtx.Err(), context.DeadlineExceeded) {
			return nil, errors.E(op, listErr, errors.KindGatewayTimeout)
		}

		// As of now, we can't reliably distinguish a true NotFound from an
		// unexpected error (https://github.com/golang/go/issues/30134), so we
		// choose the more hopeful path of NotFound (goGetErrKind defaults there,
		// while still surfacing a GitHub rate limit as 429). This lets the go
		// client fall back to `,direct` rather than treating the response as a
		// fatal 500, while we still log the underlying error for anyone digging
		// in.
		return nil, errors.E(op, listErr, goGetErrKind(stderr.String()))
	})
	if err != nil {
		return nil, nil, err
	}
	ret := sfResp.(listSFResp)
	return ret.rev, ret.versions, nil
}
