package s3

import (
	"context"
	goerr "errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// Exists implements the (./pkg/storage).Checker interface
// returning true if the module at version exists in storage.
func (s *Storage) Exists(ctx context.Context, module, version string) (bool, error) {
	const op errors.Op = "s3.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	fileKeys := []string{
		config.PackageVersionedName(module, version, "info"),
		config.PackageVersionedName(module, version, "mod"),
		config.PackageVersionedName(module, version, "zip"),
	}

	for _, key := range fileKeys {
		params := &s3.HeadObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		}

		if _, err := s.s3API.HeadObjectWithContext(ctx, params); err != nil {
			var s3Err awserr.Error
			if goerr.As(err, &s3Err) && s3Err.Code() == "NotFound" {
				return false, nil
			}
			return false, errors.E(op, err, errors.M(module), errors.V(version))
		}
	}
	return true, nil
}
