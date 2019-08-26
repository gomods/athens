package fs

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/paths"
	"github.com/spf13/afero"
)

const tokenSeparator = "|"

// Catalog implements the (./pkg/storage).Cataloger interface
// It returns a list of modules and versions contained in the storage
func (s *storageImpl) Catalog(ctx context.Context, token string, pageSize int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "fs.Catalog"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	fromModule, fromVersion, err := modVerFromToken(token)
	if err != nil {
		return nil, "", errors.E(op, err, errors.KindBadRequest)
	}

	res := make([]paths.AllPathParams, 0)
	resToken := ""
	count := pageSize

	err = afero.Walk(s.filesystem, s.rootDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".info") {
			verDir := filepath.Dir(path)
			modVer, err := filepath.Rel(s.rootDir, verDir)
			if err != nil {
				return err
			}

			m, version := filepath.Split(modVer)
			module := filepath.Clean(m)
			module = strings.Replace(module, string(os.PathSeparator), "/", -1)

			if fromModule != "" && module < fromModule { // it is ok to land on the same module
				return nil
			}

			if fromVersion != "" && version <= fromVersion { // we must skip same version
				return nil
			}

			res = append(res, paths.AllPathParams{Module: module, Version: version})
			count--
			if count == 0 {
				resToken = tokenFromModVer(module, version)
				return io.EOF
			}
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return nil, "", errors.E(op, err, errors.KindUnexpected)
	}

	return res, resToken, nil
}

func tokenFromModVer(module, version string) string {
	return module + tokenSeparator + version
}

func modVerFromToken(token string) (string, string, error) {
	const op errors.Op = "fs.Catalog"
	if token == "" {
		return "", "", nil
	}
	values := strings.Split(token, tokenSeparator)
	if len(values) < 2 {
		return "", "", errors.E(op, "Invalid token")
	}
	return values[0], values[1], nil
}
