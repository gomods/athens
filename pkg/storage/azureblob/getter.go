package azureblob

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// Info implements the (./pkg/storage).Getter interface
func (s *Storage) Info(ctx context.Context, module string, version string) ([]byte, error) {
	const op errors.Op = "azureblob.Info"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	exists, err := s.Exists(ctx, module, version)
	if err != nil {
		return nil, errors.E(op, err, errors.M(module), errors.V(version))
	}
	if !exists {
		return nil, errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}
	infoReader, err := s.cl.ReadBlob(ctx, config.PackageVersionedName(module, version, "info"))
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
func (s *Storage) GoMod(ctx context.Context, module string, vsn string) ([]byte, error) {
	const op errors.Op = "azureblob.GoMod"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	return []byte{}, errors.E(op, fmt.Errorf("Not Implemented"), errors.M(module))
}

// Zip implements the (./pkg/storage).Getter interface
func (s *Storage) Zip(ctx context.Context, module string, vsn string) (io.ReadCloser, error) {
	const op errors.Op = "azureblob.Zip"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	return nil, errors.E(op, fmt.Errorf("Not Implemented"), errors.M(module))
}
