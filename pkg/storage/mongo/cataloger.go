package mongo

import (
	"context"

	"github.com/globalsign/mgo/bson"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gomods/athens/pkg/storage"

	"github.com/gomods/athens/pkg/errors"
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

	c := s.s.DB(s.d).C(s.c)
	modules := make([]storage.Module, 0)
	err := c.Find(q).
		Select(fields).
		Sort("_id").
		Limit(pageSize).
		All(&modules)

	if err != nil {
		return nil, "", errors.E(op, err)
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
