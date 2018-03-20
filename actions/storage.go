package actions

import (
	"fmt"

	"github.com/spf13/afero"

	"github.com/gobuffalo/envy"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/fs"
)

func newStorage() (storage.Storage, error) {
	storageType := envy.Get("ATHENS_STORAGE_TYPE", "memory")
	switch storageType {
	case "memory":
		memFs := afero.NewMemMapFs()
		tmpDir, err := afero.TempDir(memFs, "inmem", "")
		if err != nil {
			return nil, fmt.Errorf("could not create temp dir for 'In Memory' storage (%s)", err)
		}
		return fs.NewStorage(tmpDir, memFs), nil
	case "disk":
		rootLocation, err := envy.MustGet("ATHENS_DISK_STORAGE_ROOT")
		if err != nil {
			return nil, fmt.Errorf("missing disk storage root (%s)", err)
		}
		return fs.NewStorage(rootLocation, afero.NewOsFs()), nil
	default:
		return nil, fmt.Errorf("storage type %s is unknown", storageType)
	}
}
