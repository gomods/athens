package mongo

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	multierror "github.com/hashicorp/go-multierror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// List lists all versions of a module
func (s *ModuleStore) List(ctx context.Context, moduleName string) ([]string, error) {
	const op errors.Op = "mongo.List"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	c := s.client.Database(s.db).Collection(s.coll)
	projection := bson.M{"version": 1, "_id": 0}
	query := bson.M{"module": moduleName}
	tctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	cursor, err := c.Find(tctx, query, &options.FindOptions{Projection: projection})
	if err != nil {
		return nil, errors.E(op, err, errors.M(moduleName))
	}
	result := make([]storage.Module, 0)
	var errs error
	for cursor.Next(ctx) {
		var module storage.Module
		if err := cursor.Decode(&module); err != nil {
			kind := errors.KindUnexpected
			if err == mongo.ErrNoDocuments {
				kind = errors.KindNotFound
			}
			errs = multierror.Append(errs, errors.E(op, err, kind))
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
