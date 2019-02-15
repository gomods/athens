package s3

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	modupl "github.com/gomods/athens/pkg/storage/module"
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
	return modupl.Exists(ctx, module, version, func(ctx context.Context, name string) (bool, error) {
		hoParams := &s3.HeadObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(name),
		}
		if _, err := s.s3API.HeadObjectWithContext(ctx, hoParams); err != nil {
			if err.(awserr.Error).Code() == s3ErrorCodeNotFound {
				return false, nil
			}

			return false, errors.E(op, err, errors.M(module))
		}

		return true, nil
	})
}
