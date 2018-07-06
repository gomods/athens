package storagetest

import (
	"testing"

	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/config/env"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/gomods/athens/pkg/storage/minio"
	"github.com/gomods/athens/pkg/storage/mongo"
	"github.com/gomods/athens/pkg/storage/rdbms"
	"github.com/spf13/afero"
)

type StorageTests struct {
	*suite.Model
	storages []storage.Backend
}

func (d *StorageTests) SetupTest() {
	ra := d.Require()

	// fs
	memFs := afero.NewOsFs()
	r, err := afero.TempDir(memFs, "", "athens-fs-storage-tests")
	ra.NoError(err)

	fsStore, err := fs.NewStorage(r, memFs)
	ra.NoError(err)

	d.storages = append(d.storages, fsStore)

	// mem
	memStore, err := mem.NewStorage()
	ra.NoError(err)

	d.storages = append(d.storages, memStore)

	// minio
	endpoint := "127.0.0.1:9000"
	bucketName := "gomods"
	accessKeyID := "minio"
	secretAccessKey := "minio123"
	minioStorage, err := minio.NewStorage(endpoint, accessKeyID, secretAccessKey, bucketName, false)
	ra.NoError(err)

	d.storages = append(d.storages, minioStorage)

	// mongo
	muri, err := env.MongoURI()
	ra.NoError(err)

	mongoStore := mongo.NewStorage(muri)
	ra.NotNil(mongoStore)
	ra.NoError(mongoStore.Connect())

	d.storages = append(d.storages, mongoStore)

	// // rdbms
	conn := d.DB
	rdbmsStore := rdbms.NewRDBMSStorageWithConn(conn)
	d.Model.SetupTest()

	d.storages = append(d.storages, rdbmsStore)
}

func TestDiskStorage(t *testing.T) {
	suite.Run(t, &StorageTests{Model: suite.NewModel()})
}
