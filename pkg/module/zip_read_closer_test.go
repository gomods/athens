package module

import (
	"path/filepath"

	"github.com/spf13/afero"
)

func (m *ModuleSuite) TestZipReadCloser() {
	const (
		root    = "testroot"
		version = "v1.0.0"
		info    = "testinfo"
		mod     = "testmod"
		zip     = "testzip"
	)
	r := m.Require()

	fs := afero.NewMemMapFs()
	gopath, err := afero.TempDir(fs, "", "athens-test")
	clean := func() error {
		return ClearFiles(fs, gopath)
	}
	r.NoError(err)
	packagePath := filepath.Join(gopath, "pkg", "mod", "cache", "download", mod, "@v")
	// create all the files the disk ref expects
	r.NoError(createAndWriteFile(fs, filepath.Join(packagePath, version+".info"), info))
	r.NoError(createAndWriteFile(fs, filepath.Join(packagePath, version+".mod"), mod))
	r.NoError(createAndWriteFile(fs, filepath.Join(packagePath, version+".zip"), zip))

	ziprc, err := fs.Open(filepath.Join(packagePath, version+".zip"))
	r.NoError(err)
	cl := &zipReadCloser{fs: fs, clean: clean, zip: ziprc}

	fInfo, err := fs.Stat(gopath)
	r.NotNil(fInfo)
	r.Nil(err)

	r.NoError(cl.Close())

	// The root dir should not exist after a clear
	fInfo, err = fs.Stat(gopath)
	r.Nil(fInfo)
	r.NotNil(err)
}

// creates filename with fs, writes data to the file, and closes the file,
//
// returns a non-nil error if anything went wrong. the file will be closed
// regardless of what this function returns
func createAndWriteFile(fs afero.Fs, filename, data string) error {
	fileHandle, err := fs.Create(filename)
	if err != nil {
		return err
	}
	defer fileHandle.Close()
	_, err = fileHandle.Write([]byte(data))
	return err
}
