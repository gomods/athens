package gcp

import (
	"context"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"google.golang.org/api/iterator"
)

// List implements the (./pkg/storage).Lister interface
// It returns a list of versions, if any, for a given module
func (s *Storage) List(ctx context.Context, module string) ([]string, error) {
	const op errors.Op = "gcp.List"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	modulePrefix := strings.TrimSuffix(module, "/") + "/@v"
	it := s.bucket.Objects(ctx, &storage.Query{Prefix: modulePrefix})
	paths := []string{}
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.E(op, err, errors.M(module))
		}
		paths = append(paths, attrs.Name)
	}

	return extractVersions(paths), nil
}

func extractVersions(paths []string) []string {
	versions := []string{}
	for _, p := range paths {
		if strings.HasSuffix(p, ".info") {
			segments := strings.Split(p, "/")
			// version should be last segment w/ .info suffix
			last := segments[len(segments)-1]
			version := strings.TrimSuffix(last, ".info")
			versions = append(versions, version)
		}
	}
	return versions
}
