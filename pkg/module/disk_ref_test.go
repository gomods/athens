package module

import (
	"io/ioutil"
)

func (m *ModuleSuite) TestDiskRefReadAndClear() {
	const (
		root    = "testroot"
		version = "v1.0.0"
		info    = "testinfo"
		mod     = "testmod"
		zip     = "testzip"
	)
	r := m.Require()

	// create a new disk ref using the filesystem
	diskRef := newDiskRef(m.fs, root)

	// ensure that reading fails, because there are no files
	ver, err := diskRef.Read()
	r.Nil(ver)
	r.NotNil(err)

	// create all the files the disk ref expects
	infoFile, err := m.fs.Create(version + ".info")
	r.NoError(err)
	defer infoFile.Close()
	_, err = infoFile.Write([]byte(info))
	r.NoError(err)

	modFile, err := m.fs.Create(version + ".mod")
	r.NoError(err)
	defer modFile.Close()
	_, err = modFile.Write([]byte(mod))
	r.NoError(err)

	srcFile, err := m.fs.Create(version + ".zip")
	r.NoError(err)
	defer srcFile.Close()
	_, err = srcFile.Write([]byte(zip))
	r.NoError(err)

	// read from the disk ref - this time it should succeed
	ver, err = diskRef.Read()
	r.NoError(err)
	r.Equal(info, string(ver.Info))
	r.Equal(mod, string(ver.Mod))
	zipBytes, err := ioutil.ReadAll(ver.Zip)
	r.NoError(err)
	r.Equal(zip, string(zipBytes))

	// clear the disk ref and expect it to fail again
	r.NoError(diskRef.Clear())
	ver, err = diskRef.Read()
	r.Nil(ver)
	r.NotNil(err)
}
