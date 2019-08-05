package minio

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

func (l *storageImpl) List(ctx context.Context, module string) ([]string, error) {
	const op errors.Op = "minio.List"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	doneCh := make(chan struct{})
	defer close(doneCh)
	searchPrefix := module + "/"
	objectCh, err := l.minioCore.ListObjectsV2(l.bucketName, searchPrefix, "", false, "", 0, "")

	if err != nil {
		return nil, errors.E(op, err, errors.M(module))
	}
	ret := []string{}
	for _, object := range objectCh.Contents {
		if object.Err != nil {
			return nil, errors.E(op, object.Err, errors.M(module))
		}
		parts := strings.Split(object.Key, "/")
		ver := parts[len(parts)-2]
		goModKey := fmt.Sprintf("%s/go.mod", l.versionLocation(module, ver))
		if goModKey == object.Key {
			ret = append(ret, ver)
		}
	}
	sort.Strings(ret)
	return ret, nil
}
