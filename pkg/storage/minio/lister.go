package minio

import (
	"context"
	"sort"
	"strings"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

func (l *storageImpl) List(ctx context.Context, module string) ([]string, error) {
	const op errors.Op = "minio.List"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	dict := make(map[string]struct{})

	doneCh := make(chan struct{})
	defer close(doneCh)
	searchPrefix := module + "/"
	objectCh := l.minioClient.ListObjectsV2(l.bucketName, searchPrefix, false, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			return nil, errors.E(op, object.Err, errors.M(module))
		}
		parts := strings.Split(object.Key, "/")
		ver := parts[len(parts)-2]
		if _, ok := dict[ver]; !ok {
			dict[ver] = struct{}{}
		}
	}

	ret := []string{}
	for ver := range dict {
		ret = append(ret, ver)
	}
	sort.Strings(ret)
	return ret, nil
}
