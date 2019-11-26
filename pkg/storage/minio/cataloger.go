package minio

import (
	"context"
	"fmt"
	"strings"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/paths"
	"github.com/minio/minio-go/v6"
)

// Catalog implements the (./pkg/storage).Cataloger interface
// It returns a list of modules and versions contained in the storage
func (s *storageImpl) Catalog(ctx context.Context, token string, pageSize int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "minio.Catalog"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	res := make([]paths.AllPathParams, 0)
	count := pageSize
	for count > 0 {
		loo, err := s.minioCore.ListObjectsV2(s.bucketName, "", token, false, "", 0, "")
		if err != nil {
			return nil, "", errors.E(op, err)
		}

		m, lastKey := fetchModsAndVersions(loo.Contents, count)

		res = append(res, m...)
		count -= len(m)
		token = lastKey
		if !loo.IsTruncated { // not truncated, there is no point in asking more
			if count > 0 { // it means we reached the end, no subsequent requests are necessary
				token = ""
			}
			break
		}
	}

	return res, token, nil
}

func fetchModsAndVersions(objects []minio.ObjectInfo, elementsNum int) ([]paths.AllPathParams, string) {
	res := make([]paths.AllPathParams, 0)
	lastKey := ""

	for _, o := range objects {
		if !strings.HasSuffix(o.Key, ".info") {
			continue
		}

		p, err := parseMinioKey(&o)
		if err != nil {
			continue
		}

		res = append(res, p)
		lastKey = o.Key

		elementsNum--
		if elementsNum == 0 {
			break
		}
	}
	return res, lastKey
}

func parseMinioKey(o *minio.ObjectInfo) (paths.AllPathParams, error) {
	const op errors.Op = "minio.parseMinioKey"

	_, m, v := extractKey(o.Key)

	if m == "" || v == "" {
		return paths.AllPathParams{}, errors.E(op, fmt.Errorf("invalid object key format %s", o.Key))
	}

	return paths.AllPathParams{Module: m, Version: v}, nil
}
