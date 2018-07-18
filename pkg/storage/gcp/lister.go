package gcp

import (
	"context"
	"strings"

	"github.com/gomods/athens/pkg/storage"
)

// List implements the (./pkg/storage).Lister interface
// It returns a list of versions, if any, for a given module
func (s *Storage) List(module string) ([]string, error) {
	ctx := context.Background()
	unfiltered, err := s.bucket.List(ctx, module)
	if err != nil {
		return nil, err
	}
	versions := make([]string, 0, 10)
	for _, n := range unfiltered {
		// kinda hacky looking at this time
		if strings.HasSuffix(n, ".info") {
			segments := strings.Split(n, "/")
			// version should be last segment w/ .info suffix
			last := segments[len(segments)-1]
			version := strings.TrimSuffix(last, ".info")
			versions = append(versions, version)
		}
	}

	if len(versions) < 1 {
		return nil, storage.ErrNotFound{Module: module}
	}
	return versions, nil
}
