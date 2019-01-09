package mongo

import (
	"context"

	"github.com/globalsign/mgo/bson"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// Exists checks for a specific version of a module
func (s *ModuleStore) Exists(ctx context.Context, module, vsn string) (bool, error) {
	var op errors.Op = "mongo.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	c := s.s.Database(s.d).Collection(s.c)
	count, err := c.Count(bson.M{"module": module, "version": vsn})
	if err != nil {
		return false, errors.E(op, errors.M(module), errors.V(vsn), err)
	}
	return count > 0, nil
}
