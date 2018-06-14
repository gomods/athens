package rdbms

import (
	"context"

	"github.com/gobuffalo/buffalo"
)

func (rd *RDBMSTestSuite) TestList() {
	c := &buffalo.DefaultContext{
		Context: context.Background(),
	}

	r := rd.Require()
	versions := []string{"v1.0.0", "v1.1.0", "v1.2.0"}
	for _, version := range versions {
		rd.storage.Save(c, module, version, mod, zip, info)
	}
	retVersions, err := rd.storage.List(module)
	r.NoError(err)
	r.Equal(versions, retVersions)
}
