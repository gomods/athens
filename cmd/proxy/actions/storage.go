package actions

import (
	"context"
	"fmt"
	"log"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/gomods/athens/pkg/storage/gcp"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/gomods/athens/pkg/storage/minio"
	"github.com/gomods/athens/pkg/storage/mongo"
	"github.com/spf13/afero"
)

// GetStorage returns storage backend based on env configuration
func GetStorage(sType string, sConf *config.StorageConfig) (storage.Backend, error) {
	const op errors.Op = "actions.GetStorage"
	log.Printf(sType)
	switch sType {
	case "memory":
		return mem.NewStorage()
	case "mongo":
		if sConf.Mongo == nil {
			return nil, errors.E(op, "Invalid Mongo Storage Configuration")
		}
		return mongo.NewStorage(sConf.Mongo)
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
	case "minio":
		if sConf.Minio == nil {
			return nil, errors.E(op, "Invalid Minio Storage Configuration")
		}
		return minio.NewStorage(sConf.Minio)
	case "gcp":
		if sConf.GCP == nil {
			return nil, errors.E(op, "Invalid GCP Storage Configuration")
		}
		if sConf.CDN == nil {
			return nil, errors.E(op, "Invalid CDN Storage Configuration")
		}
		return gcp.New(context.Background(), sConf.GCP, sConf.CDN)
	default:
		return nil, fmt.Errorf("storage type %s is unknown", sType)
	}
}
