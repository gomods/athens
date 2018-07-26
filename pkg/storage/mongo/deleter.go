package mongo

import (
	"context"

	"github.com/globalsign/mgo/bson"
	"github.com/gomods/athens/pkg/storage"
)

// Delete removes a specific version of a module
func (s *ModuleStore) Delete(ctx context.Context, module, version string) error {
	if !s.Exists(ctx, module, version) {
		return storage.ErrVersionNotFound{
			Module:  module,
			Version: version,
		}
	}
	c := s.sess.DB(athensDB).C(modulesCollection)
	return c.Remove(bson.M{"module": module, "version": version})
}
