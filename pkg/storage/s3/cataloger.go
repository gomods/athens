package s3

import (
	"context"
	"fmt"
	"strings"

	"github.com/gomods/athens/pkg/paths"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// Catalog implements the (./pkg/storage).Cataloger interface
// It returns a list of modules and versions contained in the storage
func (s *Storage) Catalog(ctx context.Context, token string, elements int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "s3.Catalog"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	queryToken := token
	res := make([]paths.AllPathParams, 0)
	for elements > 0 {
		lsParams := &s3.ListObjectsInput{
			Bucket: aws.String(s.bucket),
			Marker: &queryToken,
		}

		loo, err := s.s3API.ListObjectsWithContext(ctx, lsParams)
		if err != nil {
			return nil, "", errors.E(op, err)
		}

		m, lastKey := fetchModsAndVersions(loo.Contents, elements)
		res = append(res, m...)
		elements -= len(m)
		queryToken = lastKey

		if !*loo.IsTruncated { // not truncated, there is no point in asking more
			if elements > 0 { // it means we reached the end, no subsequent requests are necessary
				queryToken = ""
			}
			break
		}
	}
	return res, queryToken, nil
}

func fetchModsAndVersions(objects []*s3.Object, elementsNum int) ([]paths.AllPathParams, string) {
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
	fmt.Println("FEDE lastkey " + lastKey)
	return res, lastKey
}

func parseS3Key(o *s3.Object) (paths.AllPathParams, error) {
	const op errors.Op = "s3.parseS3Key"
	segments := strings.Split(*o.Key, "/")
	if len(segments) <= 0 {
		return paths.AllPathParams{}, errors.E(op, fmt.Errorf("invalid object key format %s", *o.Key))
	}
	module := segments[0]
	last := segments[len(segments)-1]
	version := strings.TrimSuffix(last, ".info")
	return paths.AllPathParams{module, version}, nil
}
