package azureblob

import (
	"context"
	"fmt"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/paths"
	"strings"
)

// Catalog implements the (./pkg/storage).Catalog interface
// It returns a list of versions, if any, for a given module
func (s *Storage) Catalog(ctx context.Context, token string, pageSize int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "azblob.Catalog"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	res := make([]paths.AllPathParams, 0)

	// one module@version consists of 3 pieces - info, mod, zip
	objCount := 3 * pageSize

	marker := azblob.Marker{&token}
	blobs, err := s.client.containerURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{
		MaxResults: int32(objCount),
	})
	if err != nil {
		return nil, "", errors.E(op, err)
	}

	nextToken := *blobs.NextMarker.Val

	for _, blob := range blobs.Segment.BlobItems {
	        if strings.HasSuffix(blob.Name, ".info") {
		        p, err := parsModVer(blob.Name)
		        if err != nil {
		                continue
		        }
		        res = append(res, p)
	        }
	}

	return res, nextToken, nil
}

func parsModVer(p string) (paths.AllPathParams, error) {
	const op errors.Op = "azureblob.parseS3Key"
	// github.com/gomods/testCatalogModule/@v/v1.2.0976.info
	m, v := config.ModuleVersionFromPath(p)

	if m == "" || v == "" {
		return paths.AllPathParams{}, errors.E(op, fmt.Errorf("invalid object key format %s", p))
	}
	return paths.AllPathParams{Module: m, Version: v}, nil
}
