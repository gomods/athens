package minio

import (
	"context"
	"fmt"
	"strings"

	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/paths"
	minio "github.com/minio/minio-go"
)

// Catalog implements the (./pkg/storage).Cataloger interface
// It returns a list of modules and versions contained in the storage
func (s *storageImpl) Catalog(ctx context.Context, token string, pageSize int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "minio.Catalog"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	queryToken := token
	res := make([]paths.AllPathParams, 0)
	count := pageSize
	for count > 0 {
		doneCh := make(chan struct{})
		defer close(doneCh)
		searchPrefix := queryToken + "/"
		loo := s.minioClient.ListObjectsV2(s.bucketName, searchPrefix, true, doneCh)

		m, lastKey := fetchModsAndVersions(loo, count)

		res = append(res, m...)
		count -= len(m)
		queryToken = lastKey
	}
	return res, queryToken, nil
}

func fetchModsAndVersions(objects <-chan minio.ObjectInfo, elementsNum int) ([]paths.AllPathParams, string) {
	res := make([]paths.AllPathParams, 0)
	lastKey := ""

	for o := range objects {
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
	m, v := config.ModuleVersionFromPath(o.Key)

	if m == "" || v == "" {
		return paths.AllPathParams{}, errors.E(op, fmt.Errorf("invalid object key format %s", o.Key))
	}
	return paths.AllPathParams{m, v}, nil
}
