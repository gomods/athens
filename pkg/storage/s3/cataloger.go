package s3

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

// Catalog implements the (./pkg/storage).Cataloger interface
// It returns a list of modules and versions contained in the storage
func (s *Storage) Catalog(ctx context.Context, token string, elements int) ([]storage.ModVer, string, error) {
	const op errors.Op = "s3.Catalog"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	lsParams := &s3.ListObjectsInput{
		Bucket: aws.String(s.bucket),
		Marker: &token,
	}

	loo, err := s.s3API.ListObjectsWithContext(ctx, lsParams)
	if err != nil {
		return nil, "", errors.E(op, err)
	}

	res, resToken := fetchModsAndVersions(loo.Contents, elements)
	return res, resToken, nil
}

func fetchModsAndVersions(objects []*s3.Object, elementsNum int) ([]storage.ModVer, string) {
	count := 0
	var res []storage.ModVer
	var token = ""

	for _, o := range objects {
		if strings.HasSuffix(*o.Key, ".info") {
			segments := strings.Split(*o.Key, "/")

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
