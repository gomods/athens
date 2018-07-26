package mongo

import (
	"context"

	"github.com/globalsign/mgo/bson"
)

// Exists checks for a specific version of a module
func (s *ModuleStore) Exists(ctx context.Context, module, vsn string) bool {
	c := s.sess.DB(athensDB).C(modulesCollection)
	count, err := c.Find(bson.M{"module": module, "version": vsn}).Count()
	return err == nil && count > 0
}
