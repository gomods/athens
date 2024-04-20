package s3

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	moduploader "github.com/gomods/athens/pkg/storage/module"
)

// Save implements the (github.com/gomods/athens/pkg/storage).Saver interface.
func (s *Storage) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	const op errors.Op = "s3.Save"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	err := moduploader.Upload(ctx, module, version, bytes.NewReader(info), bytes.NewReader(mod), zip, s.upload, s.timeout)
	// TODO: take out lease on the /list file and add the version to it
	//
	// Do that only after module source+metadata is uploaded
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	return nil
}

func (s *Storage) upload(ctx context.Context, path, contentType string, stream io.Reader) error {
	const op errors.Op = "s3.upload"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	upParams := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(path),
		Body:        stream,
		ContentType: aws.String(contentType),
	}

	if _, err := s.uploader.Upload(ctx, upParams); err != nil {
		return errors.E(op, err)
	}

	return nil
}
