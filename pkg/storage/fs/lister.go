package fs

import (
	"context"
	"os"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/spf13/afero"
)

func (l *storageImpl) List(ctx context.Context, module string) ([]string, error) {
	const op errors.Op = "fs.List"
	_, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	loc := l.moduleLocation(module)
	fileInfos, err := afero.ReadDir(l.filesystem, loc)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}

		return nil, errors.E(op, errors.M(module), err, errors.KindUnexpected)
	}
	ret := []string{}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			ret = append(ret, fileInfo.Name())
		}
	}
	return ret, nil
}
