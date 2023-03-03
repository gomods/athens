package s3

import (
	"context"
	errs "errors"

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

	files := []string{"info", "mod", "zip"}
	cancelingCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	errChan := make(chan error, len(files))
	defer close(errChan)
	for _, file := range files {
		go func(file string) {
			_, err := s.s3API.HeadObjectWithContext(
				cancelingCtx,
				&s3.HeadObjectInput{
					Bucket: aws.String(s.bucket),
					Key:    aws.String(config.PackageVersionedName(module, version, file)),
				})
			errChan <- err
		}(file)
	}
	for _ = range files {
		err := <-errChan
		if err != nil {
			var aerr awserr.Error
			if errs.As(err, &aerr) && aerr.Code() == s3.ErrCodeNoSuchKey {
				return false, nil
			}
			return false, err
		}
	}
	return true, nil
}
