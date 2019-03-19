package s3

import (
	"context"

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
		Prefix: aws.String(config.PackageVersionedName(module, version, "")),
	}

	loo, err := s.s3API.ListObjectsWithContext(ctx, lsParams)
	if err != nil {
		return false, errors.E(op, err, errors.M(module), errors.V(version))
	}

	return len(loo.Contents) == 3, nil
}
