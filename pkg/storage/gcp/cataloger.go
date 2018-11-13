package gcp

import (
	"context"
	"strings"

	"github.com/gomods/athens/pkg/paths"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// Catalog implements the (./pkg/storage).Catalog interface
// It returns a list of versions, if any, for a given module
func (s *Storage) Catalog(ctx context.Context, token string, elements int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "gcp.Catalog"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	catalog, token, err := s.bucket.Catalog(ctx, token, elements)
	if err != nil {
		return nil, "", errors.E(op, err)
	}
	res, resToken := fetchModsAndVersions(catalog, elements)
	return res, resToken, nil
}

func fetchModsAndVersions(catalog []string, elementsNum int) ([]paths.AllPathParams, string) {
	count := 0
	var res []paths.AllPathParams
	var token = ""

	for _, p := range catalog {
		if strings.HasSuffix(p, ".info") {
			segments := strings.Split(p, "/")

			if len(segments) <= 0 {
				continue
			}
			module := segments[0]
			last := segments[len(segments)-1]
			version := strings.TrimSuffix(last, ".info")
			res = append(res, paths.AllPathParams{module, version})
			count++
		}

		if count == elementsNum {
			break
		}
	}

	return res, token
}
