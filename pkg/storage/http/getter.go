package http

import (
	"context"
	"io"
	"io/ioutil"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// Info implements storage.Getter
func (s *ModuleStore) Info(ctx context.Context, module, vsn string) ([]byte, error) {
	const op errors.Op = "http.Info"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	body, err := s.fetch(ctx, s.moduleRoot(module)+vsn+".info")
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer body.Close()

	return ioutil.ReadAll(body)
}

// GoMod implements storage.Getter
func (s *ModuleStore) GoMod(ctx context.Context, module, vsn string) ([]byte, error) {
	const op errors.Op = "http.GoMod"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	body, err := s.fetch(ctx, s.moduleRoot(module)+vsn+".mod")
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer body.Close()

	return ioutil.ReadAll(body)
}

// Zip implements storage.Getter
func (s *ModuleStore) Zip(ctx context.Context, module, vsn string) (io.ReadCloser, error) {
	const op errors.Op = "http.Zip"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	return s.fetch(ctx, s.moduleRoot(module)+vsn+".zip")
}
