package mongo

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gomods/athens/pkg/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/hashicorp/go-multierror"
)

// Catalog implements the (./pkg/storage).Cataloger interface
// It returns a list of modules and versions contained in the storage
func (s *ModuleStore) Catalog(ctx context.Context, token string, pageSize int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "mongo.Catalog"
	q := bson.E{}
	if token != "" {
		q = bson.E{Key: "_id", Value: bson.M{"$gt": token}}
	}

	projection := bson.M{"module": 1, "version": 1}

	c := s.client.Database(s.db).Collection(s.coll)

	modules := make([]storage.Module, 0)
	cursor, err := c.Find(context.Background(), q, &options.FindOptions{Projection: projection})

	if err != nil {
		return nil, "", errors.E(op, err)
	}

	var errs error
	for cursor.Next(context.Background()) {
		var module storage.Module
		elem := &bson.D{}
		if err := cursor.Decode(module); err != nil {
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
