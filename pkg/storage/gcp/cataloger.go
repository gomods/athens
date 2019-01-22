package gcp

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/paths"
	"google.golang.org/api/iterator"
)

// Catalog implements the (./pkg/storage).Catalog interface
// It returns a list of versions, if any, for a given module
func (s *Storage) Catalog(ctx context.Context, token string, pageSize int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "gcp.Catalog"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	res := make([]paths.AllPathParams, 0)
	var resToken string
	count := pageSize

	for count > 0 {
		var catalog []string
		var err error
		catalog, resToken, err = nextPage(ctx, s.bucket, token, 3*count)
		if err != nil {
			return nil, "", errors.E(op, err)
		}
		pathsAndVers := fetchModsAndVersions(catalog)
		res = append(res, pathsAndVers...)
		count -= len(pathsAndVers)

		if resToken == "" { // meaning we reached the end
			break
		}
	}
	return res, resToken, nil
}

func nextPage(ctx context.Context, bkt *storage.BucketHandle, token string, pageSize int) ([]string, string, error) {
	it := bkt.Objects(ctx, nil)
	p := iterator.NewPager(it, pageSize, token)

	attrs := make([]*storage.ObjectAttrs, 0)
	nextToken, err := p.NextPage(&attrs)
	if err != nil {
		return nil, "", err
	}

	res := []string{}
	for _, attr := range attrs {
		res = append(res, attr.Name)
	}
	return res, nextToken, nil
}

func fetchModsAndVersions(catalog []string) []paths.AllPathParams {
	res := make([]paths.AllPathParams, 0)
	for _, p := range catalog {
		if !strings.HasSuffix(p, ".info") {
			continue
		}
		p, err := parseGcpKey(p)
		if err != nil {
			continue
		}
		res = append(res, p)
	}
	return res
}

func parseGcpKey(p string) (paths.AllPathParams, error) {
	const op errors.Op = "gcp.parseS3Key"
	// github.com/gomods/testCatalogModule/@v/v1.2.0976.info
	m, v := config.ModuleVersionFromPath(p)

	if m == "" || v == "" {
		return paths.AllPathParams{}, errors.E(op, fmt.Errorf("invalid object key format %s", p))
	}
	return paths.AllPathParams{m, v}, nil
}
