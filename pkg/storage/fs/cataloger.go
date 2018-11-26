package fs

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/paths"
	"github.com/spf13/afero"

	"github.com/gomods/athens/pkg/errors"
)

// Catalog implements the (./pkg/storage).Cataloger interface
// It returns a list of modules and versions contained in the storage
func (s *storageImpl) Catalog(ctx context.Context, token string, elements int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "fs.Catalog"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	moduleInfos, err := afero.ReadDir(s.filesystem, s.rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", nil
		}
		return nil, "", errors.E(op, err, errors.KindUnexpected)
	}

	res := make([]paths.AllPathParams, 0)
	fromModule, fromVersion, err := modVerFromToken(token)
	if err != nil {
		return nil, "", errors.E(op, err, errors.KindBadRequest)
	}

	count := elements
	sortFsSlice(moduleInfos)
	for _, moduleInfo := range moduleInfos {
		if fromModule != "" && moduleInfo.Name() < fromModule { // is it ok to land on the same module
			continue
		}

		if moduleInfo.IsDir() {
			moduleDir := filepath.Join(s.rootDir, moduleInfo.Name())
			versionInfos, err := afero.ReadDir(s.filesystem, moduleDir)
			if err != nil && os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, "", errors.E(op, err, errors.KindUnexpected)
			}
			sortFsSlice(versionInfos)
			for _, versionInfo := range versionInfos {
				if fromVersion != "" && versionInfo.Name() <= fromVersion { // we must skip same version
					continue
				}
				res = append(res, paths.AllPathParams{moduleInfo.Name(), versionInfo.Name()})
				count--
				if elements > 0 && count == 0 {
					return res, tokenFromModVer(moduleInfo.Name(), versionInfo.Name()), nil
				}
			}
		}
	}

	return res, "", nil
}

func sortFsSlice(toSort []os.FileInfo) {
	sort.Slice(toSort, func(i, j int) bool {
		if toSort[i].Name() < toSort[j].Name() {
			return true
		}
		return false
	})
}

func tokenFromModVer(module, version string) string {
	return module + "|" + version
}

func modVerFromToken(token string) (string, string, error) {
	const op errors.Op = "fs.Catalog"
	if token == "" {
		return "", "", nil
	}
	values := strings.Split(token, "|")
	if len(values) < 2 {
		return "", "", errors.E(op, "Invalid token")
	}
	return values[0], values[1], nil
}
