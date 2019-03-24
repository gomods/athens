package mongo

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"go.mongodb.org/mongo-driver/bson"
)

// Exists checks for a specific version of a module
func (s *ModuleStore) Exists(ctx context.Context, module, vsn string) (bool, error) {
	var op errors.Op = "mongo.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	c := s.client.Database(s.db).Collection(s.coll)
	tctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	count, err := c.CountDocuments(tctx, bson.M{"module": module, "version": vsn})
	if err != nil {
		return false, errors.E(op, errors.M(module), errors.V(vsn), err)
	}
	return count > 0, nil
}
