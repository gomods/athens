package mongo

import (
	"context"
	"io"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observability"
	"github.com/gomods/athens/pkg/storage"
)

// Save stores a module in mongo storage.
func (s *ModuleStore) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	const op errors.Op = "storage.mongo.Save"
	ctx, span := observability.StartSpan(ctx, op.String())
	defer span.End()

	zipName := s.gridFileName(module, version)
	fs := s.s.DB(s.d).GridFS("fs")
	f, err := fs.Create(zipName)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	defer f.Close()

	_, err = io.Copy(f, zip) // check number of bytes written?
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	m := &storage.Module{
		Module:  module,
		Version: version,
		Mod:     mod,
		Info:    info,
	}

	c := s.s.DB(s.d).C(s.c)
	err = c.Insert(m)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	return nil
}
