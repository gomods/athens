package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/observ"
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
	found := make(map[string]struct{}, 3)
	err := s.s3API.ListObjectsPagesWithContext(ctx, lsParams, func(loo *s3.ListObjectsOutput, lastPage bool) bool {
		for _, o := range loo.Contents {
			if _, exists := found[*o.Key]; exists {
				log.EntryFromContext(ctx).Warnf("duplicate key in prefix %q: %q", *lsParams.Prefix, *o.Key)
				continue
			}
			if *o.Key == config.PackageVersionedName(module, version, "info") ||
				*o.Key == config.PackageVersionedName(module, version, "mod") ||
				*o.Key == config.PackageVersionedName(module, version, "zip") {
				found[*o.Key] = struct{}{}
			}
		}
		return len(found) < 3
	})

	if err != nil {
		return false, errors.E(op, err, errors.M(module), errors.V(version))
	}
	return len(found) == 3, nil
}
