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
	searchPrefix := module
	objectCh := l.minioClient.ListObjectsV2(l.bucketName, searchPrefix, true, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			return nil, errors.E(op, object.Err, errors.M(module))
		}
		parts := strings.Split(object.Key, "/")
		modPart := parts[len(parts)-1]
		ver := stripExtension(modPart)
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

func stripExtension(modPart string) string {
	modPart = strings.TrimSuffix(modPart, ".mod")
	modPart = strings.TrimSuffix(modPart, ".info")
	modPart = strings.TrimSuffix(modPart, ".zip")
	return modPart
}
