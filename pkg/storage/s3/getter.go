package s3

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
		var nsk *types.NoSuchKey
		if errors.AsErr(err, &nsk) {
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
		var nsk *types.NoSuchKey
		if errors.AsErr(err, &nsk) {
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
		var nsk *types.NoSuchKey
		if errors.AsErr(err, &nsk) {
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

	if s.cdnEndpoint != nil {
		return s.openFromCDN(ctx, path)
	}

	getParams := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}

	goo, err := s.s3API.GetObject(ctx, getParams)
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.AsErr(err, &nsk) {
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

func (s *Storage) openFromCDN(ctx context.Context, path string) (storage.SizeReadCloser, error) {
	const op errors.Op = "s3.open"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.cdnObjectURL(path), nil)
	if err != nil {
		return nil, errors.E(op, err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			return nil, errors.E(op, errors.KindNotFound)
		}
		return nil, errors.E(op, fmt.Errorf("unexpected CDN status %d for %s", resp.StatusCode, path))
	}
	var size int64
	if resp.ContentLength > 0 {
		size = resp.ContentLength
	}
	return storage.NewSizer(resp.Body, size), nil
}
