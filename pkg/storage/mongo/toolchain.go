package mongo

import (
	"context"
	"io"
	"regexp"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/internal/toolchain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *ModuleStore) toolchains() *toolchain.Store {
	return &toolchain.Store{Objects: &releaseObjects{s}}
}

func (s *ModuleStore) Archive(ctx context.Context, source, filename string) (storage.ArchiveInfo, storage.SizeReadCloser, error) {
	return s.toolchains().Archive(ctx, source, filename)
}

func (s *ModuleStore) SaveArchive(ctx context.Context, source string, info storage.ArchiveInfo, body io.Reader) error {
	return s.toolchains().SaveArchive(ctx, source, info, body)
}

func (s *ModuleStore) ListArchives(ctx context.Context, source string) ([]storage.ArchiveInfo, error) {
	return s.toolchains().ListArchives(ctx, source)
}

func (s *ModuleStore) DeleteArchive(ctx context.Context, source, filename string) error {
	return s.toolchains().DeleteArchive(ctx, source, filename)
}

func (s *ModuleStore) Releases(ctx context.Context, source string) (*storage.ReleaseList, error) {
	return s.toolchains().Releases(ctx, source)
}

func (s *ModuleStore) SaveReleases(ctx context.Context, source string, list *storage.ReleaseList) error {
	return s.toolchains().SaveReleases(ctx, source, list)
}

type (
	releaseObjects struct{ store *ModuleStore }
	releaseObject  struct {
		Key    string        `bson:"_id"`
		FileID bson.ObjectID `bson:"file_id"`
		Size   int64         `bson:"size"`
	}
)

func (o *releaseObjects) collection() *mongo.Collection {
	return o.store.client.Database(o.store.db).Collection(o.store.coll + "_toolchains")
}

func (o *releaseObjects) bucket() *mongo.GridFSBucket {
	return o.store.client.Database(o.store.db).GridFSBucket(options.GridFSBucket().SetName(o.store.coll + "_toolchain_bodies"))
}

func (o *releaseObjects) Open(ctx context.Context, key string) (storage.SizeReadCloser, error) {
	var obj releaseObject
	err := o.collection().FindOne(ctx, bson.M{"_id": key}).Decode(&obj)
	if err != nil {
		if errors.IsErr(err, mongo.ErrNoDocuments) {
			return nil, errors.E("mongo.Archive", err, errors.KindNotFound)
		}
		return nil, err
	}
	body, err := o.bucket().OpenDownloadStream(ctx, obj.FileID)
	if err != nil {
		return nil, err
	}
	return storage.NewSizer(body, obj.Size), nil
}

func (o *releaseObjects) Create(ctx context.Context, key string, body io.ReadSeeker, size int64) error {
	bucket := o.bucket()
	upload, err := bucket.OpenUploadStream(ctx, key)
	if err != nil {
		return err
	}
	n, err := io.Copy(upload, body)
	if err != nil || n != size {
		_ = upload.Abort()
		if err != nil {
			return err
		}
		return errors.E("mongo.SaveArchive", "incomplete object")
	}
	if err := upload.Close(); err != nil {
		return err
	}
	obj := releaseObject{Key: key, FileID: upload.FileID.(bson.ObjectID), Size: size}
	_, err = o.collection().InsertOne(ctx, obj)
	if err != nil {
		_ = bucket.Delete(ctx, obj.FileID)
		if mongo.IsDuplicateKeyError(err) {
			return errors.E("mongo.SaveArchive", err, errors.KindAlreadyExists)
		}
	}
	return err
}

func (o *releaseObjects) Keys(ctx context.Context, prefix string) ([]string, error) {
	cursor, err := o.collection().Find(ctx, bson.M{"_id": bson.M{"$regex": "^" + regexp.QuoteMeta(prefix)}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var keys []string
	for cursor.Next(ctx) {
		var obj releaseObject
		if err := cursor.Decode(&obj); err != nil {
			return nil, err
		}
		keys = append(keys, obj.Key)
	}
	return keys, cursor.Err()
}

func (o *releaseObjects) Delete(ctx context.Context, key string) error {
	var obj releaseObject
	err := o.collection().FindOneAndDelete(ctx, bson.M{"_id": key}).Decode(&obj)
	if errors.IsErr(err, mongo.ErrNoDocuments) {
		return errors.E("mongo.DeleteArchive", err, errors.KindNotFound)
	}
	if err != nil {
		return err
	}
	return o.bucket().Delete(ctx, obj.FileID)
}
