package modulestorage

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/gomods/athens/pkg/storage/minio"
	"github.com/gomods/athens/pkg/storage/mongo"
	"github.com/gomods/athens/pkg/storage/rdbms"
	"github.com/stretchr/testify/require"
)

func BenchmarkStorageList(b *testing.B) {
	module, version := "athens_module", "1.0.1"
	zip, info, mod := []byte("zip_data"), []byte("info"), []byte("mod_info")

	for _, store := range getStores(b) {

		backend := store.Storage()

		require.NoError(b, backend.Save(context.Background(), module, version, mod, bytes.NewReader(zip), info), "Save for storage %s failed", backend)

		b.Run(fmt.Sprintf("listing module backend %s", store.StorageHumanReadableName()), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				backend.List(context.Background(), module)
			}
		})

		require.NoError(b, store.Cleanup())
	}
}

func BenchmarkStorageSave(b *testing.B) {
	module, version := "athens_module", "1.0.1"
	zip, info, mod := []byte("zip_data"), []byte("info"), []byte("mod_info")

	for _, store := range getStores(b) {

		backend := store.Storage()

		b.Run(fmt.Sprintf("save module backend %s", store.StorageHumanReadableName()), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := backend.Save(context.Background(), fmt.Sprintf("%s-%d", module, i), version, mod, bytes.NewReader(zip), info)
				require.NoError(b, err)
			}
		})

		require.NoError(b, store.Cleanup())
	}
}

func BenchmarkStorageDelete(b *testing.B) {
	module, version := "athens_module", "1.0.1"
	zip, info, mod := []byte("zip_data"), []byte("info"), []byte("mod_info")

	for _, store := range getStores(b) {

		backend := store.Storage()

		for i := 0; i < b.N; i++ {
			err := backend.Save(context.Background(), fmt.Sprintf("del-%s-%d", module, i), version, mod, bytes.NewReader(zip), info)
			require.NoError(b, err, "save for storage %s module: %s failed", backend, i)
		}

		b.Run(fmt.Sprintf("delete module backend %s", store.StorageHumanReadableName()), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				backend.Delete(context.Background(), fmt.Sprintf("del-%s-%d", module, i), version)
			}
		})

		require.NoError(b, store.Cleanup())
	}
}

func BenchmarkStorageExists(b *testing.B) {
	module, version := "athens_module", "1.0.1"
	zip, info, mod := []byte("zip_data"), []byte("info"), []byte("mod_info")

	for _, store := range getStores(b) {

		backend := store.Storage()

		for i := 0; i < b.N; i++ {
			err := backend.Save(context.Background(), fmt.Sprintf("%s-%d", module, i), version, mod, bytes.NewReader(zip), info)
			require.NoError(b, err, "exists for storage %s failed", backend)
		}

		b.Run(fmt.Sprintf("exists module backend %s", store.StorageHumanReadableName()), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				backend.Exists(context.Background(), fmt.Sprintf("%s-%d", module, i), version)
			}
		})

		b.Run(fmt.Sprintf("non existent module backend %s", store.StorageHumanReadableName()), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				backend.Exists(context.Background(), fmt.Sprintf("invaid-%s-%d", module, i), version)
			}
		})
		require.NoError(b, store.Cleanup())
	}
}

func getStores(b *testing.B) []storage.TestSuite {
	var stores []storage.TestSuite

	//TODO: create the instance without model or TestSuite
	model := suite.NewModel()
	fsStore, err := fs.NewTestSuite(model)
	require.NoError(b, err, "couldn't create filesystem store")
	stores = append(stores, fsStore)

	mongoStore, err := mongo.NewTestSuite(model)
	require.NoError(b, err, "couldn't create mongo store")
	stores = append(stores, mongoStore)

	rdbmsStore, err := rdbms.NewTestSuite(model)
	require.NoError(b, err, "couldn't create mongo store")
	stores = append(stores, rdbmsStore)

	memStore, err := mem.NewTestSuite(model)
	require.NoError(b, err)
	stores = append(stores, memStore)

	minioStore, err := minio.NewTestSuite(model)
	require.NoError(b, err)
	stores = append(stores, minioStore)

	return stores
}
