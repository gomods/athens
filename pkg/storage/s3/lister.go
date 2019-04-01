package s3

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// List implements the (./pkg/storage).Lister interface
// It returns a list of versions, if any, for a given module
func (s *Storage) List(ctx context.Context, module string) ([]string, error) {
	const op errors.Op = "s3.List"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	modulePrefix := strings.TrimSuffix(module, "/") + "/@v"
	lsParams := &s3.ListObjectsInput{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(modulePrefix),
	}

	loo, err := s.s3API.ListObjectsWithContext(ctx, lsParams)
	if err != nil {
		return nil, errors.E(op, err, errors.M(module))
	}

	return extractVersions(loo.Contents), nil
}

func extractVersions(objects []*s3.Object) []string {
	var versions []string

	for _, o := range objects {
		if strings.HasSuffix(*o.Key, ".info") {
			segments := strings.Split(*o.Key, "/")

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
