package azureblob

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/paths"
)

// Catalog implements the (./pkg/storage).Catalog interface.
// It returns a list of versions, if any, for a given module.
func (s *Storage) Catalog(ctx context.Context, token string, pageSize int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "azblob.Catalog"

	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	res := make([]paths.AllPathParams, 0)

	// one module@version consists of 3 pieces - info, mod, zip
	objCount := int32(3 * pageSize)

	opts := &azblob.ListBlobsFlatOptions{
		MaxResults: &objCount,
	}

	if token != "" {
		opts.Marker = &token
	}

	pager := s.client.client.NewListBlobsFlatPager(s.client.containerName, opts)

	if !pager.More() {
		return res, "", nil
	}

	resp, err := pager.NextPage(ctx)
	if err != nil {
		return nil, "", errors.E(op, err)
	}

	var nextToken string

	if resp.NextMarker != nil {
		nextToken = *resp.NextMarker
	}

	for _, blob := range resp.Segment.BlobItems {
		if strings.HasSuffix(*blob.Name, ".info") {
			p, err := parsModVer(*blob.Name)
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
