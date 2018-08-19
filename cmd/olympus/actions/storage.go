package actions

import (
	"fmt"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/gomods/athens/pkg/storage/mongo"
	"github.com/spf13/afero"
)

// GetStorage returns storage.Backend implementation
func GetStorage(sType string, sConf *config.StorageConfig) (storage.Backend, error) {
	const op errors.Op = "actions.GetStorage"
	switch sType {
	case "memory":
		return mem.NewStorage()
	case "disk":
		if sConf.Disk == nil {
			return nil, errors.E(op, "Invalid Disk Storage Configuration")
		}
		rootLocation := sConf.Disk.RootPath
		s, err := fs.NewStorage(rootLocation, afero.NewOsFs())
		if err != nil {
			errStr := fmt.Sprintf("could not create new storage from os fs (%s)", err)
			return nil, errors.E(op, errStr)
		}
		return s, nil
	case "mongo":
		if sConf.Mongo == nil {
			return nil, errors.E(op, "Invalid Mongo Storage Configuration")
		}
		mongoURI := sConf.Mongo.URL
		mongoTimeout := config.TimeoutDuration(sConf.Mongo.Timeout)
		return mongo.NewStorage(mongoURI, mongoTimeout)
	default:
		errStr := fmt.Sprintf("storage type %s is unknown", sType)
		return nil, errors.E(op, errStr)
	}
}
