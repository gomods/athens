package s3

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"io/ioutil"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// Info implements the (./pkg/storage).Getter interface
func (s *Storage) Info(ctx context.Context, module, version string) ([]byte, error) {
	const op errors.Op = "s3.Info"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	exists, err := s.Exists(ctx, module, version)
	if err != nil {
		return nil, errors.E(op, err, errors.M(module), errors.V(version))
	}
	if !exists {
		return nil, errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}

	infoReader, err := s.open(ctx, config.PackageVersionedName(module, version, "info"))
	if err != nil {
		return nil, errors.E(op, err, errors.M(module), errors.V(version))
	}
	defer infoReader.Close()

	infoBytes, err := ioutil.ReadAll(infoReader)
	if err != nil {
		return nil, errors.E(op, err, errors.M(module), errors.V(version))
	}
	return infoBytes, nil
}

// GoMod implements the (./pkg/storage).Getter interface
func (s *Storage) GoMod(ctx context.Context, module, version string) ([]byte, error) {
	const op errors.Op = "s3.GoMod"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	exists, err := s.Exists(ctx, module, version)
	if err != nil {
		return nil, errors.E(op, err, errors.M(module), errors.V(version))
	}
	if !exists {
		return nil, errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}

	modReader, err := s.open(ctx, config.PackageVersionedName(module, version, "mod"))
	if err != nil {
		return nil, errors.E(op, err, errors.M(module), errors.V(version))
	}
	defer modReader.Close()

	modBytes, err := ioutil.ReadAll(modReader)
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("could not get new reader for mod file: %s", err), errors.M(module), errors.V(version))
	}

	return modBytes, nil
}

// Zip implements the (./pkg/storage).Getter interface
func (s *Storage) Zip(ctx context.Context, module, version string) (io.ReadCloser, error) {
	const op errors.Op = "s3.Zip"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	exists, err := s.Exists(ctx, module, version)
	if err != nil {
		return nil, errors.E(op, err, errors.M(module), errors.V(version))
	}
	if !exists {
		return nil, errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}

	zipReader, err := s.open(ctx, config.PackageVersionedName(module, version, "zip"))
	if err != nil {
		return nil, errors.E(op, err, errors.M(module), errors.V(version))
	}

	return zipReader, nil
}

func (s *Storage) open(ctx context.Context, path string) (io.ReadCloser, error) {
	const op errors.Op = "s3.open"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	getParams := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}

	goo, err := s.s3API.GetObjectWithContext(ctx, getParams)
	if err != nil {
		return nil, errors.E(op, err, errors.M(path))
	}

	return goo.Body, nil
}
