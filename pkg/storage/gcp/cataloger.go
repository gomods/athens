package gcp

import (
	"context"
	"fmt"
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
	res := make([]paths.AllPathParams, 0)
	var resToken string

	for elements > 0 {
		var catalog []string
		var err error
		catalog, resToken, err = s.bucket.Catalog(ctx, token, elements)
		if err != nil {
			return nil, "", errors.E(op, err)
		}
		pathsAndVers := fetchModsAndVersions(catalog)
		res = append(res, pathsAndVers...)
		elements -= len(pathsAndVers)

		if resToken == "" { // meaning we reached the end
			break
		}
	}
	return res, resToken, nil
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
	const op errors.Op = "gcp.parseGcpKey"
	segments := strings.Split(p, "/")
	if len(segments) <= 0 {
		return paths.AllPathParams{}, errors.E(op, fmt.Errorf("invalid object key format %s", p))
	}
	module := segments[0]
	last := segments[len(segments)-1]
	version := strings.TrimSuffix(last, ".info")
	return paths.AllPathParams{module, version}, nil
}
