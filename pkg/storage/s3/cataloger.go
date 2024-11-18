package s3

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/paths"
)

// Catalog implements the (./pkg/storage).Cataloger interface.
// It returns a list of modules and versions contained in the storage.
func (s *Storage) Catalog(ctx context.Context, token string, pageSize int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "s3.Catalog"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	queryToken := token
	res := make([]paths.AllPathParams, 0)
	count := pageSize
	for count > 0 {
		lsParams := &s3.ListObjectsV2Input{
			Bucket:     aws.String(s.bucket),
			StartAfter: &queryToken,
		}

		loo, err := s.s3API.ListObjectsV2(ctx, lsParams)
		if err != nil {
			return nil, "", errors.E(op, err)
		}

		m, lastKey := fetchModsAndVersions(loo.Contents, count)

		res = append(res, m...)
		count -= len(m)
		queryToken = lastKey

		if !*loo.IsTruncated { // not truncated, there is no point in asking more
			if count > 0 { // it means we reached the end, no subsequent requests are necessary
				queryToken = ""
			}
			break
		}
	}
	return res, queryToken, nil
}

func fetchModsAndVersions(objects []types.Object, elementsNum int) ([]paths.AllPathParams, string) {
	res := make([]paths.AllPathParams, 0)
	lastKey := ""
	for _, o := range objects {
		if !strings.HasSuffix(*o.Key, ".info") {
			continue
		}
		p, err := parseS3Key(o)
		if err != nil {
			continue
		}

		res = append(res, p)
		lastKey = *o.Key

		elementsNum--
		if elementsNum == 0 {
			break
		}
	}
	return res, lastKey
}

func parseS3Key(o types.Object) (paths.AllPathParams, error) {
	const op errors.Op = "s3.parseS3Key"
	m, v := config.ModuleVersionFromPath(*o.Key)

	if m == "" || v == "" {
		return paths.AllPathParams{}, errors.E(op, fmt.Errorf("invalid object key format %s", *o.Key))
	}
	return paths.AllPathParams{Module: m, Version: v}, nil
}
