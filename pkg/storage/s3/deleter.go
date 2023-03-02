package s3

import (
	"context"
	errs "errors"
	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	modupl "github.com/gomods/athens/pkg/storage/module"
)

// Delete implements the (./pkg/storage).Deleter interface and
// removes a version of a module from storage. Returning ErrNotFound
// if the version does not exist.
func (s *Storage) Delete(ctx context.Context, module, version string) (err error) {
	const op errors.Op = "s3.Delete"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	if err = modupl.Delete(ctx, module, version, s.remove, s.timeout); err != nil {
		var aerr awserr.Error
		if errs.As(err, &aerr) && aerr.Code() == s3.ErrCodeNoSuchKey {
			return errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
		}
	}
	return err
}

func (s *Storage) remove(ctx context.Context, path string) error {
	const op errors.Op = "s3.Delete"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	delParams := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}

	if _, err := s.s3API.DeleteObjectWithContext(ctx, delParams); err != nil {
		return errors.E(op, err)
	}

	return nil
}
