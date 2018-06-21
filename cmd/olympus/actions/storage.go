package actions

import (
	"fmt"

	"github.com/gobuffalo/envy"
	"github.com/gomods/athens/pkg/env"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/gomods/athens/pkg/storage/mongo"
	"github.com/gomods/athens/pkg/storage/rdbms"
	"github.com/spf13/afero"
)

// GetStorage returns a storage.Backend implementation based on the
// ATHENS_STORAGE_TYPE env var
func GetStorage() (storage.Backend, error) {
	storageType := envy.Get("ATHENS_STORAGE_TYPE", "memory")
	switch storageType {
	case "memory":
		return mem.NewStorage()
	case "disk":
		rootLocation, err := envy.MustGet("ATHENS_DISK_STORAGE_ROOT")
		if err != nil {
			return nil, fmt.Errorf("missing disk storage root (%s)", err)
		}
		s := fs.NewStorage(rootLocation, afero.NewOsFs())
		return storage.NoOpBackend(s), nil
	case "mongo":
		mongoDeets, err := env.ForMongo()
		if err != nil {
			return nil, fmt.Errorf("mongo configuration error (%s)", err)
		}
		return mongo.NewStorage(mongoDeets)
	case "postgres", "sqlite", "cockroach", "mysql":
		connectionName, err := envy.MustGet("ATHENS_RDBMS_STORAGE_NAME")
		if err != nil {
			return nil, fmt.Errorf("missing rdbms connectionName (%s)", err)
		}
		return rdbms.NewRDBMSStorage(connectionName)
	default:
		return nil, fmt.Errorf("storage type %s is unknown", storageType)
	}
}
