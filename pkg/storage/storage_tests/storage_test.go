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

type TestSuites struct {
	*suite.Model
	storages []storage.TestSuite
}

func (d *TestSuites) SetupTest() {
	ra := d.Require()

	//
	fsTests, err := fs.NewTestSuite(d.Model)
	ra.NoError(err)
	d.storages = append(d.storages, fsTests)

	// mem
	memStore, err := mem.NewTestSuite(d.Model)
	ra.NoError(err)
	d.storages = append(d.storages, memStore)

	// minio
	minioStorage, err := minio.NewTestSuite(d.Model)
	ra.NoError(err)
	d.storages = append(d.storages, minioStorage)

	// mongo
	mongoStore, err := mongo.NewTestSuite(d.Model)
	ra.NoError(err)
	d.storages = append(d.storages, mongoStore)

	// rdbms
	rdbmsStore, err := rdbms.NewTestSuite(d.Model)
	d.Model.SetupTest()
	d.storages = append(d.storages, rdbmsStore)
}

func TestDiskStorage(t *testing.T) {
	suite.Run(t, &TestSuites{Model: suite.NewModel()})
}

func (d *TestSuites) TestStorages() {
	for _, store := range d.storages {
		d.testNotFound(store)

		// TODO: more tests to come
	}
}

func (d *TestSuites) testNotFound(ts storage.TestSuite) {
	_, err := ts.Storage().Get("some", "unknown")
	d.Require().Equal(true, storage.IsNotFoundError(err), "Invalid error type for %s: %#v", ts.StorageHumanReadableName(), err)
}
