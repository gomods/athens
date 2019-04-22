package mongo

import (
	"context"
	"fmt"
	"io"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Save stores a module in mongo storage.
func (s *ModuleStore) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	const op errors.Op = "mongo.Save"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	exists, err := s.Exists(ctx, module, version)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	if exists {
		return errors.E(op, "already exists", errors.M(module), errors.V(version), errors.KindAlreadyExists)
	}

	zipName := s.gridFileName(module, version)
	db := s.client.Database(s.db)
	bucket, err := gridfs.NewBucket(db, options.GridFSBucket())
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	uStream, err := bucket.OpenUploadStream(zipName, options.GridFSUpload())
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	defer uStream.Close()

	numBytesWritten, err := io.Copy(uStream, zip)

	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	if numBytesWritten <= 0 {
		e := fmt.Errorf("copied %d bytes to Mongo GridFS", numBytesWritten)
		return errors.E(op, e, errors.M(module), errors.V(version))
	}

	m := &storage.Module{
		Module:  module,
		Version: version,
		Mod:     mod,
		Info:    info,
	}

	c := s.client.Database(s.db).Collection(s.coll)
	tctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	_, err = c.InsertOne(tctx, m, options.InsertOne().SetBypassDocumentValidation(false))
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	return nil
}
