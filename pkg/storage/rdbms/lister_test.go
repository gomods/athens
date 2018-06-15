package rdbms

import (
	"context"

	"github.com/bketelsen/buffet"

	"github.com/gobuffalo/buffalo"
)

func (rd *RDBMSTestSuite) TestList() {
	c := &buffalo.DefaultContext{
		Context: context.Background(),
	}
	sp := buffet.SpanFromContext(c)
	sp.SetOperationName("test.storage.rdbms.List")
	defer sp.Finish()

	r := rd.Require()
	versions := []string{"v1.0.0", "v1.1.0", "v1.2.0"}
	for _, version := range versions {
		rd.storage.Save(c, module, version, mod, zip, info)
	}
	retVersions, err := rd.storage.List(module)
	r.NoError(err)
	r.Equal(versions, retVersions)
}
