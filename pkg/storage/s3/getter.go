package s3

import (
	"context"
	errs "errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

// Info implements the (./pkg/storage).Getter interface.
func (s *Storage) Info(ctx context.Context, module, version string) ([]byte, error) {
	const op errors.Op = "s3.Info"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	infoReader, err := s.open(ctx, config.PackageVersionedName(module, version, "info"))
	if err != nil {
		var aerr awserr.Error
		if errs.As(err, &aerr) && aerr.Code() == s3.ErrCodeNoSuchKey {
			return nil, errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
		}
		return nil, errors.E(op, err, errors.M(module), errors.V(version))
	}
	defer func() { _ = infoReader.Close() }()

	infoBytes, err := io.ReadAll(infoReader)
	if err != nil {
		return nil, errors.E(op, err, errors.M(module), errors.V(version))
	}
	return infoBytes, nil
}

// GoMod implements the (./pkg/storage).Getter interface.
func (s *Storage) GoMod(ctx context.Context, module, version string) ([]byte, error) {
	const op errors.Op = "s3.GoMod"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	modReader, err := s.open(ctx, config.PackageVersionedName(module, version, "mod"))
	if err != nil {
		var aerr awserr.Error
		if errs.As(err, &aerr) && aerr.Code() == s3.ErrCodeNoSuchKey {
			return nil, errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
		}
		return nil, errors.E(op, err, errors.M(module), errors.V(version))
	}
	defer func() { _ = modReader.Close() }()

	modBytes, err := io.ReadAll(modReader)
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("could not get new reader for mod file: %w", err), errors.M(module), errors.V(version))
	}

	return modBytes, nil
}

// Zip implements the (./pkg/storage).Getter interface.
func (s *Storage) Zip(ctx context.Context, module, version string) (storage.SizeReadCloser, error) {
	const op errors.Op = "s3.Zip"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	zipReader, err := s.open(ctx, config.PackageVersionedName(module, version, "zip"))
	if err != nil {
		var aerr awserr.Error
		if errs.As(err, &aerr) && aerr.Code() == s3.ErrCodeNoSuchKey {
			return nil, errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
		}
		return nil, errors.E(op, err, errors.M(module), errors.V(version))
	}

	return zipReader, nil
}

func (s *Storage) open(ctx context.Context, path string) (storage.SizeReadCloser, error) {
	const op errors.Op = "s3.open"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	getParams := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}

	goo, err := s.s3API.GetObjectWithContext(ctx, getParams)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchKey {
			return nil, errors.E(op, errors.KindNotFound)
		}
		return nil, errors.E(op, err)
	}
	var size int64
	if goo.ContentLength != nil {
		size = *goo.ContentLength
	}
	return storage.NewSizer(goo.Body, size), nil
}
