package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

const (
	s3ErrorCodeNotFound = "NotFound"
)

// Exists implements the (./pkg/storage).Checker interface
// returning true if the module at version exists in storage
func (s *Storage) Exists(ctx context.Context, module, version string) (bool, error) {
	const op errors.Op = "s3.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	lsParams := &s3.ListObjectsInput{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(fmt.Sprintf("%s/@v", module)),
	}
	var count int
	err := s.s3API.ListObjectsPagesWithContext(ctx, lsParams, func(loo *s3.ListObjectsOutput, lastpage bool) bool {
		for _, o := range loo.Contents {
			// sane assumption: no duplicate keys.
			switch *o.Key {
			case config.PackageVersionedName(module, version, "info"):
				count++
			case config.PackageVersionedName(module, version, "mod"):
				count++
			case config.PackageVersionedName(module, version, "zip"):
				count++
			}
		}
		// stop paging if already found all files or this is the last page
		if count == 3 || lastpage {
			return false
		}
		return true
	})

	if err != nil {
		return false, errors.E(op, err, errors.M(module), errors.V(version))
	}

	return count == 3, nil
}
