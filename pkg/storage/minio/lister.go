package minio

import (
	"context"
	"sort"
	"strings"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observability"
)

func (l *storageImpl) List(ctx context.Context, module string) ([]string, error) {
	const op errors.Op = "storage.minio.List"
	ctx, span := observability.StartSpan(ctx, op.String())
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
	if len(dict) == 0 {
		return ret, errors.E(op, errors.M(module), errors.KindNotFound)
	}

	for ver := range dict {
		ret = append(ret, ver)
	}
	sort.Strings(ret)
	return ret, nil
}
