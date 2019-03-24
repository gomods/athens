package mongo

import (
	"context"

	"github.com/globalsign/mgo"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/hashicorp/go-multierror"
)

// List lists all versions of a module
func (s *ModuleStore) List(ctx context.Context, module string) ([]string, error) {
	const op errors.Op = "mongo.List"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	c := s.client.Database(s.db).Collection(s.coll)
	projection := bson.M{"version": 1}
	query := bson.E{Key: "module", Value: module}
	tctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	cursor, err := c.Find(tctx, query, &options.FindOptions{Projection: projection})
	if err != nil {
		kind := errors.KindUnexpected
		if err == mgo.ErrNotFound {
			kind = errors.KindNotFound
		}
		return nil, errors.E(op, kind, errors.M(module), err)
	}
	result := make([]storage.Module, 0)
	var errs error;
	for cursor.Next(ctx) {
		var module storage.Module
		if err := cursor.Decode(module); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			result = append(result, module)
		}
	}

	versions := make([]string, len(result))
	for i, r := range result {
		versions[i] = r.Version
	}

	return versions, nil
}
