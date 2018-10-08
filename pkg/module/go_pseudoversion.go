package module

import (
	"context"
	"encoding/json"
	"os/exec"
	"regexp"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/paths"
	"github.com/spf13/afero"
)

type goListResult struct {
	Path    string `json:"Path"`
	Version string `json:"Version"`
	Time    string `json:"String"`
}

// PseudoVersionFromHash returns the go mod pseudoversion associated to the given commit hash used as version
func PseudoVersionFromHash(ctx context.Context, fs afero.Fs, gobinary, mod, ver string) (string, error) {
	const op errors.Op = "module.PseudoVersionFromHash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	if IsSemVersion(ver) {
		return ver, nil
	}

	fullURI := paths.FmtModVer(mod, ver)
	cmd := exec.Command(gobinary, "list", "-m", "-json", fullURI)

	tmpRoot, err := afero.TempDir(fs, "", "pseudover")
	modPath, err := setupModRepo(fs, tmpRoot, mod, ver)
	defer ClearFiles(fs, tmpRoot)

	if err != nil {
		return "", errors.E(op, err)
	}

	cmd.Env = PrepareEnv(tmpRoot)
	cmd.Dir = modPath

	o, err := cmd.Output()
	if err != nil {
		return "", errors.E(op, err)
	}

	var r goListResult
	err = json.Unmarshal(o, &r)
	if err != nil {
		return "", errors.E(op, err)
	}
	return r.Version, nil
}

// IsSemVersion tells whether the passed string respects the semantic version pattern
func IsSemVersion(ver string) bool {
	res, _ := regexp.Match("v\\d+\\.\\d+.\\d+", []byte(ver))
	return res
}
