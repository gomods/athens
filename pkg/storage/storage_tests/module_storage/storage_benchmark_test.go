package modulestorage

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/gomods/athens/pkg/storage/mongo"
	"github.com/stretchr/testify/require"
)

func BenchmarkStorageList(b *testing.B) {
	module, version := "athens_module", "1.0.1"
	zip, info, mod := []byte("zip_data"), []byte("info"), []byte("mod_info")

	for _, store := range getStores(b) {

		backend := store.Storage()
		store.Cleanup()

		require.NoError(b, backend.Save(context.Background(), module, version, mod, bytes.NewReader(zip), info), "Save for storage %s failed", backend)

		for i := 0; i < b.N; i++ {
			b.Run(fmt.Sprintf("listing module: backend %s", store.StorageHumanReadableName()), func(b *testing.B) {
				backend.List(context.Background(), module)
			})
		}

		store.Cleanup()
	}
}

func getStores(b *testing.B) []storage.TestSuite {
	var stores []storage.TestSuite

	//TODO: create the instance without model
	model := suite.NewModel()
	fsTests, err := fs.NewTestSuite(model)
	require.NoError(b, err, "couldn't create filesystem store")
	stores = append(stores, fsTests)

	mongoStore, err := mongo.NewTestSuite(model)
	require.NoError(b, err, "couldn't create mongo store")
	stores = append(stores, mongoStore)

	return stores
}
