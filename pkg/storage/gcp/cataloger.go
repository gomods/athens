package gcp

import (
	"context"
	"strings"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

// Catalog implements the (./pkg/storage).Catalog interface
// It returns a list of versions, if any, for a given module
func (s *Storage) Catalog(ctx context.Context, token string, elements int) ([]storage.ModVer, string, error) {
	const op errors.Op = "gcp.Catalog"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	paths, token, err := s.bucket.Catalog(ctx, token, elements)
	if err != nil {
		return nil, "", errors.E(op, err)
	}
	res, resToken := fetchModsAndVersions(paths, elements)
	return res, resToken, nil
}

func fetchModsAndVersions(paths []string, elementsNum int) ([]storage.ModVer, string) {
	count := 0
	var res []storage.ModVer
	var token = ""

	for _, p := range paths {
		if strings.HasSuffix(p, ".info") {
			segments := strings.Split(p, "/")

			if len(segments) <= 0 {
				continue
			}
			module := segments[0]
			last := segments[len(segments)-1]
			version := strings.TrimSuffix(last, ".info")
			res = append(res, storage.ModVer{module, version})
			count++
		}

		if count == elementsNum {
			break
		}
	}

	return res, token
}
