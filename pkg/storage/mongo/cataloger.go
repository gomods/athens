package mongo

import (
	"context"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gomods/athens/pkg/storage"
	"github.com/hashicorp/go-multierror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Catalog implements the (./pkg/storage).Cataloger interface
// It returns a list of modules and versions contained in the storage
func (s *ModuleStore) Catalog(ctx context.Context, token string, pageSize int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "mongo.Catalog"
	q := bson.M{}
	if token != "" {
		t, err := primitive.ObjectIDFromHex(token)
		if err == nil {
			q = bson.M{"_id": bson.M{"$gt": t}}
		}
	}

	projection := bson.M{"module": 1, "version": 1}
	sort := bson.M{"_id": 1}

	c := s.client.Database(s.db).Collection(s.coll)

	tctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	modules := make([]storage.Module, 0)
	findOptions := options.Find().SetProjection(projection).SetSort(sort).SetLimit(int64(pageSize))
	cursor, err := c.Find(tctx, q, findOptions)

	if err != nil {
		return nil, "", errors.E(op, err)
	}

	var errs error
	for cursor.Next(ctx) {
		var module storage.Module
		if err := cursor.Decode(&module); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			modules = append(modules, module)
		}
	}

	// If there are 0 results, return empty results without an error
	if len(modules) == 0 {
		return nil, "", nil
	}

	var versions = make([]paths.AllPathParams, len(modules))
	for i := range modules {
		versions[i].Module = modules[i].Module
		versions[i].Version = modules[i].Version
	}

	var next = modules[len(modules)-1].ID.Hex()
	if len(modules) < pageSize {
		return versions, "", nil
	}
	return versions, next, nil
}
