package s3

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	moduploader "github.com/gomods/athens/pkg/storage/module"
)

// Save implements the (github.com/gomods/athens/pkg/storage).Saver interface.
func (s *Storage) Save(ctx context.Context, module, version string, mod []byte, zip storage.Zip, info []byte) error {
	const op errors.Op = "s3.Save"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	zipStream := moduploader.Stream{zip.Reader, zip.Size}
	err := moduploader.Upload(ctx, module, version, moduploader.NewStreamFromBytes(info), moduploader.NewStreamFromBytes(mod), zipStream, s.upload, s.timeout)
	// TODO: take out lease on the /list file and add the version to it
	//
	// Do that only after module source+metadata is uploaded
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	return nil
}

func (s *Storage) upload(ctx context.Context, path, contentType string, stream moduploader.Stream) error {
	const op errors.Op = "s3.upload"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	upParams := &s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(path),
		Body:        stream.Stream,
		ContentType: aws.String(contentType),
	}

	if _, err := s.uploader.UploadWithContext(ctx, upParams); err != nil {
		return errors.E(op, err)
	}

	return nil
}
