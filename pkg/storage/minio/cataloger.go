package minio

import (
	"context"
	"fmt"
	"strings"

	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/paths"
	"github.com/minio/minio-go"
)

// Catalog implements the (./pkg/storage).Cataloger interface
// It returns a list of modules and versions contained in the storage
func (s *storageImpl) Catalog(ctx context.Context, token string, pageSize int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "minio.Catalog"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	queryToken := token
	res := make([]paths.AllPathParams, 0)
	doneCh := make(chan struct{})
	defer close(doneCh)
	loo := s.minioClient.ListObjectsV2(s.bucketName, token, true, doneCh)

	m, lastKey := fetchModsAndVersions(loo, pageSize)

	res = append(res, m...)
	queryToken = lastKey

	return res, queryToken, nil
}

func fetchModsAndVersions(objects <-chan minio.ObjectInfo, elementsNum int) ([]paths.AllPathParams, string) {
	res := make([]paths.AllPathParams, 0)
	lastKey := ""

	for o := range objects {
		fmt.Println(o.Key)
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
	parts := strings.Split(o.Key, "/")
	fmt.Println(parts)
	v := parts[len(parts)-1]
	fmt.Println(v)

	m := strings.Replace(o.Key, v + "/.info", "", -1)
	if m == "" || v == "" {
		return paths.AllPathParams{}, errors.E(op, fmt.Errorf("invalid object key format %s", o.Key))
	}
	return paths.AllPathParams{m, v}, nil
}
