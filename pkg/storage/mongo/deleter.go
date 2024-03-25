package mongo

import (
	"context"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Delete removes a specific version of a module.
func (s *ModuleStore) Delete(ctx context.Context, module, version string) error {
	const op errors.Op = "mongo.Delete"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	exists, err := s.Exists(ctx, module, version)
	if err != nil {
		return errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}
	if !exists {
		return errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}

	db := s.client.Database(s.db)
	c := db.Collection(s.coll)
	bucket, err := gridfs.NewBucket(db, &options.BucketOptions{})
	if err != nil {
		return errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
	}

	filter := bson.D{bson.E{Key: "filename", Value: s.gridFileName(module, version)}}

	cursor, err := bucket.Find(filter)
	if err != nil {
		return errors.E(op, errors.M(module), errors.V(version), err)
	}

	var x bson.D
	for cursor.Next(ctx) {
		_ = cursor.Decode(&x)
	}
	b, err := bson.Marshal(x)
	if err != nil {
		return errors.E(op, errors.M(module), errors.V(version), err)
	}

	if err = bucket.Delete(bson.Raw(b).Lookup("_id").ObjectID()); err != nil {
		kind := errors.KindUnexpected
		if errors.IsErr(err, gridfs.ErrFileNotFound) {
			kind = errors.KindNotFound
		}
		return errors.E(op, err, kind, errors.M(module), errors.V(version))
	}

	tctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	_, err = c.DeleteOne(tctx, bson.M{"module": module, "version": version})
	if err != nil {
		return errors.E(op, err, errors.KindNotFound, errors.M(module), errors.V(version))
	}
	return nil
}
