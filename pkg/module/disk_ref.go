package module

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/gomods/athens/pkg/storage"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// DiskRef is a Ref implementation for modules on disk. It is not safe to use concurrently.
//
// Do not create this struct directly. use newDiskRef
type diskRef struct {
	root         string
	fs           afero.Fs
	version      string
	filesToClose []afero.File
}

func newDiskRef(fs afero.Fs, root, version string) *diskRef {
	return &diskRef{
		fs:      fs,
		root:    root,
		version: version,
	}
}

// Clear is the Ref interface implementation. It deletes all module data from disk
//
// You should always call this function after you fetch a module into a DiskRef
func (d *diskRef) Clear() error {
	var errors error
	for _, file := range d.filesToClose {
		if err := file.Close(); err != nil {
			multierror.Append(errors, err)
		}
	}
	if err := d.fs.RemoveAll(d.root); err != nil {
		multierror.Append(errors, err)
	}
	return errors
}

// read is the Ref interface implementation.
func (d *diskRef) Read() (*storage.Version, error) {
	ver := &storage.Version{}

	infoFile, err := d.fs.Open(filepath.Join(d.root, fmt.Sprintf("%s.info", d.version)))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	d.filesToClose = append(d.filesToClose, infoFile)

	info, err := ioutil.ReadAll(infoFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ver.Info = info

	modFile, err := d.fs.Open(filepath.Join(d.root, fmt.Sprintf("%s.mod", d.version)))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	d.filesToClose = append(d.filesToClose, modFile)
	mod, err := ioutil.ReadAll(modFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ver.Mod = mod

	sourceFile, err := d.fs.Open(filepath.Join(d.root, fmt.Sprintf("%s.zip", d.version)))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	d.filesToClose = append(d.filesToClose, sourceFile)
	ver.Zip = sourceFile

	return ver, nil
}
