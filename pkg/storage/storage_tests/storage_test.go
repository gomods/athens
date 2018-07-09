package storagetest

import (
	"testing"

	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/gomods/athens/pkg/storage/minio"
	"github.com/gomods/athens/pkg/storage/mongo"
	"github.com/gomods/athens/pkg/storage/rdbms"
)

type StorageTests struct {
	*suite.Model
	storages []storage.StorageTest
}

func (d *StorageTests) SetupTest() {
	ra := d.Require()

	//
	fsTests, err := fs.NewStorageTest(d.Model)
	ra.NoError(err)
	d.storages = append(d.storages, fsTests)

	// mem
	memStore, err := mem.NewStorageTest(d.Model)
	ra.NoError(err)
	d.storages = append(d.storages, memStore)

	// minio
	minioStorage, err := minio.NewStorageTest(d.Model)
	ra.NoError(err)
	d.storages = append(d.storages, minioStorage)

	// mongo
	mongoStore, err := mongo.NewStorageTest(d.Model)
	ra.NoError(err)
	d.storages = append(d.storages, mongoStore)

	// rdbms
	rdbmsStore, err := rdbms.NewStorageTest(d.Model)
	d.Model.SetupTest()
	d.storages = append(d.storages, rdbmsStore)
}

func TestDiskStorage(t *testing.T) {
	suite.Run(t, &StorageTests{Model: suite.NewModel()})
}

func (d *StorageTests) TestNotFound() {
	for _, store := range d.storages {
		store.TestNotFound()
	}
}
