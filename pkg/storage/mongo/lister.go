package mongo

import (
	"context"
	"strings"

	"github.com/globalsign/mgo/bson"
	"github.com/gomods/athens/pkg/storage"
)

// List lists all versions of a module
func (s *ModuleStore) List(ctx context.Context, module string) ([]string, error) {
	c := s.sess.DB(athensDB).C(modulesCollection)
	result := make([]storage.Module, 0)
	err := c.Find(bson.M{"module": module}).All(&result)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			err = storage.ErrNotFound{Module: module}
		}
		return nil, err
	}

	versions := make([]string, len(result))
	for i, r := range result {
		versions[i] = r.Version
	}

	return versions, nil
}
