package download

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
)

type Lister func(string) (*storage.RevInfo, []string, error)

type listResp struct {
	Path     string
	Version  string
	Versions []string `json:",omitempty"`
	Time     time.Time
}

// NewGoLister bla
func NewGoLister(goBinPath string, fs afero.Fs) Lister {
	return func(mod string) (*storage.RevInfo, []string, error) {
		const op errors.Op = "download.Lister"
		hackyPath, err := afero.TempDir(fs, "", "hackymod")
		if err != nil {
			return nil, nil, errors.E(op, err)
		}
		defer fs.RemoveAll(hackyPath)
		err = module.Dummy(fs, hackyPath)
		if err != nil {
			return nil, nil, errors.E(op, err)
		}

		cmd := exec.Command(
			goBinPath,
			"list", "-m", "-versions", "-json",
			config.FmtModVer(mod, "latest"),
		)
		cmd.Dir = hackyPath
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		gopath, err := afero.TempDir(fs, "", "athens")
		if err != nil {
			return nil, nil, errors.E(op, err)
		}
		defer module.ClearFiles(fs, gopath)
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
}
