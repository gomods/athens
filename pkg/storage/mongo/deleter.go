package mongo

import (
	"github.com/globalsign/mgo/bson"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// Delete removes a specific version of a module
func (s *ModuleStore) Delete(ctx observ.ProxyContext, module, version string) error {
	const op errors.Op = "mongo.Delete"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	exists, err := s.Exists(ctx, module, version)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	if !exists {
		return errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}

	db := s.s.DB(s.d)
	c := db.C(s.c)
	err = db.GridFS("fs").Remove(s.gridFileName(module, version))
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	err = c.Remove(bson.M{"module": module, "version": version})
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	return nil
}
