package mongo

import (
	"context"
	"io"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Info implements storage.Getter
func (s *ModuleStore) Info(ctx context.Context, module, vsn string) ([]byte, error) {
	const op errors.Op = "mongo.Info"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	c := s.client.Database(s.db).Collection(s.coll)

	result := &storage.Module{}

	tctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	queryResult := c.FindOne(tctx, bson.M{"module": module, "version": vsn})
	if queryErr := queryResult.Err(); queryErr != nil {
		return nil, errors.E(op, queryErr, errors.M(module), errors.V(vsn))
	}

	if err := queryResult.Decode(&result); err != nil {
		kind := errors.KindUnexpected
		if err == mongo.ErrNoDocuments {
			kind = errors.KindNotFound
		}
		return nil, errors.E(op, err, kind, errors.M(module), errors.V(vsn))
	}

	return result.Info, nil
}

// GoMod implements storage.Getter
func (s *ModuleStore) GoMod(ctx context.Context, module, vsn string) ([]byte, error) {
	const op errors.Op = "mongo.GoMod"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	c := s.client.Database(s.db).Collection(s.coll)
	result := &storage.Module{}
	tctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	queryResult := c.FindOne(tctx, bson.M{"module": module, "version": vsn})
	if queryErr := queryResult.Err(); queryErr != nil {
		return nil, errors.E(op, queryErr, errors.M(module), errors.V(vsn))
	}

	if err := queryResult.Decode(result); err != nil {
		kind := errors.KindUnexpected
		if err == mongo.ErrNoDocuments {
			kind = errors.KindNotFound
		}
		return nil, errors.E(op, err, kind, errors.M(module), errors.V(vsn))
	}

	return result.Mod, nil
}

// Zip implements storage.Getter
func (s *ModuleStore) Zip(ctx context.Context, module, vsn string) (io.ReadCloser, error) {
	const op errors.Op = "mongo.Zip"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	zipName := s.gridFileName(module, vsn)
	db := s.client.Database(s.db)
	bucket, err := gridfs.NewBucket(db, &options.BucketOptions{})
	if err != nil {
		return nil, errors.E(op, err, errors.M(module), errors.V(vsn))
	}

	dStream, err := bucket.OpenDownloadStreamByName(zipName, options.GridFSName())
	if err != nil {
		kind := errors.KindUnexpected
		if err == gridfs.ErrFileNotFound {
			kind = errors.KindNotFound
		}
		return nil, errors.E(op, err, kind, errors.M(module), errors.V(vsn))
	}

	return dStream, nil
}
