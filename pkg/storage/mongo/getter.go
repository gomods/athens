package mongo

import (
	"context"
	"io"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

// Info implements storage.Getter
func (s *ModuleStore) Info(ctx context.Context, module, vsn string) ([]byte, error) {
	const op errors.Op = "mongo.Info"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	c := s.s.Database(s.d).Collection(s.c)
	result := &storage.Module{}
	err := c.FindOne(bson.M{"module": module, "version": vsn}).Decode(result)
	if err != nil {
		kind := errors.KindUnexpected
		if err == mgo.ErrNotFound {
			kind = errors.KindNotFound
		}
		return nil, errors.E(op, kind, errors.M(module), errors.V(vsn), err)
	}

	return result.Info, nil
}

// GoMod implements storage.Getter
func (s *ModuleStore) GoMod(ctx context.Context, module, vsn string) ([]byte, error) {
	const op errors.Op = "mongo.GoMod"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	c := s.s.Database(s.d).Collection(s.c)
	result := &storage.Module{}
	err := c.FindOne(bson.M{"module": module, "version": vsn}).Decode(result)
	if err != nil {
		kind := errors.KindUnexpected
		if err == mgo.ErrNotFound {
			kind = errors.KindNotFound
		}
		return nil, errors.E(op, kind, errors.M(module), errors.V(vsn), err)
	}

	return result.Mod, nil
}

// Zip implements storage.Getter
func (s *ModuleStore) Zip(ctx context.Context, module, vsn string) (io.ReadCloser, error) {
	const op errors.Op = "mongo.Zip"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	zipName := s.gridFileName(module, vsn)
	db := s.s.Database(s.d)
	fs := s.s.Database(s.d).GridFS("fs")
	f, err := fs.Open(zipName)
	if err != nil {
		kind := errors.KindUnexpected
		if err == mgo.ErrNotFound {
			kind = errors.KindNotFound
		}
		return nil, errors.E(op, err, kind, errors.M(module), errors.V(vsn))
	}

	return f, nil
}
