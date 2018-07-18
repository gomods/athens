package module

import (
	"fmt"
	"io/ioutil"

	"github.com/gomods/athens/pkg/storage"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// DiskRef is a Ref implementation for modules on disk. It is not safe to use concurrently.
//
// Do not create this struct directly. use newDiskRef
type diskRef struct {
	root    string
	fs      afero.Fs
	version string
	closer  func()
}

func newDiskRef(fs afero.Fs, root string) *diskRef {
	return &diskRef{fs: fs, root: root, closer: func() {}}
}

// Clear is the Ref interface implementation. It deletes all module data from disk
//
// You should always call this function after you fetch a module into a DiskRef
func (d *diskRef) Clear() error {
	d.closer()
	return d.fs.RemoveAll(d.root)
}

// read is the Ref interface implementation.
func (d *diskRef) Read() (*storage.Version, error) {
	ver := &storage.Version{}

	infoFile, err := d.fs.Open(fmt.Sprintf("%s.info", d.version))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer infoFile.Close()
	info, err := ioutil.ReadAll(infoFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ver.Info = info

	modFile, err := d.fs.Open(fmt.Sprintf("%s.mod", d.version))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer modFile.Close()
	mod, err := ioutil.ReadAll(modFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ver.Mod = mod

	sourceFile, err := d.fs.Open(fmt.Sprintf("%s.zip", d.version))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	d.closer = func() { sourceFile.Close() }
	ver.Zip = sourceFile

	return ver, nil
}
