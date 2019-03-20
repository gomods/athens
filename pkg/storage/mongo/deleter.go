package mongo

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Delete removes a specific version of a module
func (s *ModuleStore) Delete(ctx context.Context, module, version string) error {
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

	db := s.client.Database(s.db)
	c := db.Collection(s.coll)
	bucket, err := gridfs.NewBucket(db, &options.BucketOptions{})
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	err = bucket.Delete(s.gridFileName(module, version))
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	result, err := c.DeleteOne(context.Background(), bson.M{"module": module, "version": version})
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	return nil
}
