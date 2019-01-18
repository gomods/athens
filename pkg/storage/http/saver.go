package http

import (
	"bytes"
	"context"
	"io"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	modupl "github.com/gomods/athens/pkg/storage/module"
)

// Save stores a module in mongo storage.
func (s *ModuleStore) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	const op errors.Op = "http.Save"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	err := modupl.Upload(ctx, module, version, bytes.NewReader(info), bytes.NewReader(mod), zip, s.upload, s.timeout)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	return nil
}
