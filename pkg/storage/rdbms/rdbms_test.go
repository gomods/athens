package rdbms

import (
	"io/ioutil"
)

func (rd *RDBMSTests) TestGetSaveListRoundTrip() {
	r := rd.Require()
	rd.storage.Save(baseURL, module, version, mod, zip)
	listedVersions, err := rd.storage.List(baseURL, module)
	r.NoError(err)
	r.Equal(1, len(listedVersions))
	retVersion := listedVersions[0]
	r.Equal(version, retVersion)
	gotten, err := rd.storage.Get(baseURL, module, version)
	r.NoError(err)
	defer gotten.Zip.Close()
	r.Equal(version, gotten.RevInfo.Version)
	r.Equal(version, gotten.RevInfo.Name)
	r.Equal(version, gotten.RevInfo.Short)
	// TODO: test the time
	r.Equal(gotten.Mod, mod)
	zipContent, err := ioutil.ReadAll(gotten.Zip)
	r.NoError(err)
	r.Equal(zipContent, zip)
}

func (rd *RDBMSTests) TestNewRDBMSStorage() {
	r := rd.Require()
	e := "development"
	getterSaver := NewRDBMSStorage(e)
	getterSaver.Connect()

	r.NotNil(getterSaver.conn)
	r.Equal(getterSaver.e, e)
}
