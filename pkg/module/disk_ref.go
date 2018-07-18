package module

import (
	"fmt"
	"io/ioutil"

	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
)

// DiskRef is a Ref implementation for modules on disk
type DiskRef struct {
	root    string
	fs      afero.Fs
	version string
}

func newDiskRef(fs afero.Fs, root string) *DiskRef {
	return &DiskRef{fs: fs, root: root}
}

// Clear is the Ref interface implementation. It deletes all module data from disk
//
// You should always call this function after you fetch a module into a DiskRef
func (d *DiskRef) Clear() error {
	return d.fs.RemoveAll(d.root)
}

func (d *DiskRef) Read() (*storage.Version, error) {
	ver := &storage.Version{}

	infoFile, err := d.fs.Open(fmt.Sprintf("%s.info", d.version))
	if err != nil {
		return nil, err
	}
	defer infoFile.Close()

	modFile, err := d.fs.Open(fmt.Sprintf("%s.mod", d.version))
	if err != nil {
		return nil, err
	}
	defer modFile.Close()

	sourceFile, err := d.fs.Open(fmt.Sprintf("%s.zip", d.version))
	if err != nil {
		return nil, err
	}
	defer sourceFile.Close()

	info, err := ioutil.ReadAll(infoFile)
	if err != nil {
		return nil, err
	}
}
