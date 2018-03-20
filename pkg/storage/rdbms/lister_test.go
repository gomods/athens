package rdbms

func (rd *RDBMSTests) TestList() {
	r := rd.Require()
	versions := []string{"v1.0.0", "v1.1.0", "v1.2.0"}
	for _, version := range versions {
		rd.storage.Save(baseURL, module, version, mod, zip)
	}
	retVersions, err := rd.storage.List(baseURL, module)
	r.NoError(err)
	r.Equal(versions, retVersions)
}