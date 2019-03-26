package azureblob

import (
	"context"
	"strings"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// List implements the (./pkg/storage).Lister interface
// It returns a list of versions, if any, for a given module
func (s *Storage) List(ctx context.Context, module string) ([]string, error) {
	const op errors.Op = "azureblob.List"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	blobnames, err := s.client.ListBlobs(ctx, module)
	if err != nil {
		return nil, errors.E(op, err, errors.M(module))
	}
	return extractVersions(blobnames), nil
}

func extractVersions(blobnames []string) []string {
	var versions []string

	for _, b := range blobnames {
		if strings.HasSuffix(b, ".info") {
			segments := strings.Split(b, "/")

			if len(segments) == 0 {
				continue
			}
			// version should be last segment w/ .info suffix
			last := segments[len(segments)-1]
			version := strings.TrimSuffix(last, ".info")
			versions = append(versions, version)
		}
	}
	return versions
}
