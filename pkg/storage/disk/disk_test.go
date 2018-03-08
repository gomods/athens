package disk

import (
	"path/filepath"
)

func (d *DiskTests) TestLocationFuncs() {
	r := d.Require()
	storage := d.storage.(*storageImpl)
	moduleLoc := storage.moduleDiskLocation(baseURL, module)
	r.Equal(filepath.Join(d.rootDir, baseURL, module), moduleLoc)
	versionedLoc := storage.versionDiskLocation(baseURL, module, version)
	r.Equal(filepath.Join(d.rootDir, baseURL, module, version), versionedLoc)
}
