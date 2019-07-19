package module

import (
	"io"
	"os"

	"github.com/gomods/athens/pkg/errors"
	"github.com/spf13/afero"
)

type zipReadCloser struct {
	zip   io.ReadCloser
	fs    afero.Fs
	clean runtimeClean
}

// Close closes the zip file handle and clears up disk space used by the underlying disk ref
// It is the caller's responsibility to call this method to free up utilized disk space
func (rc *zipReadCloser) Close() error {
	rc.zip.Close()
	if rc.clean == nil {
		return nil
	}
	return rc.clean()
}

func (rc *zipReadCloser) Read(p []byte) (n int, err error) {
	return rc.zip.Read(p)
}

// ClearFiles deletes all data from the given fs at path root
// This function must be called when zip is closed to cleanup the entire GOPATH created by the diskref
func ClearFiles(fs afero.Fs, root string) error {
	const op errors.Op = "module.ClearFiles"
	// This is required because vgo ensures dependencies are read-only
	// See https://github.com/golang/go/issues/24111 and
	// https://go-review.googlesource.com/c/vgo/+/96978
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return fs.Chmod(path, 0770)
	}
	err := afero.Walk(fs, root, walkFn)
	if err != nil {
		return errors.E(op, err)
	}
	err = fs.RemoveAll(root)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}
