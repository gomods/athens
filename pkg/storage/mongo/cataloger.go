package mongo

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gomods/athens/pkg/storage"
	"github.com/mongodb/mongo-go-driver/bson"
)

// Catalog implements the (./pkg/storage).Cataloger interface
// It returns a list of modules and versions contained in the storage
func (s *ModuleStore) Catalog(ctx context.Context, token string, pageSize int) ([]paths.AllPathParams, string, error) {
	const op errors.Op = "mongo.Catalog"
	q := bson.M{}
	if token != "" {
		q = bson.M{"_id": bson.M{"$gt": bson.ObjectIdHex(token)}}
	}

	fields := bson.M{"module": 1, "version": 1}

	compositeQ := bson.D{q, fields}

	c := s.client.Database(s.db).Collection(s.coll)

	modules := make([]storage.Module, 0)
	cursor, err := c.Find(compositeQ)
	// * Currently the driver doesn't have any of the below provisions so maybe adding the
	// field projection along with the query will work
	//
	// Select(fields).
	// Sort("_id").
	// Limit(pageSize).
	// All(&modules)
	// *

	if err != nil {
		return nil, "", errors.E(op, err)
	}
	for cursor.Next() {
		var module storage.Module
		bytes, err := cursor.DecodeBytes()
		if err == nil {
			err = bson.Unmarshal(bytes, &module)
			if err == nil {
				modules = append(modules, module)
			}
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
