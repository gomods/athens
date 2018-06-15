package mongo

import (
	"context"

	"github.com/bketelsen/buffet"

	"github.com/gobuffalo/buffalo"
)

func (m *MongoTests) TestList() {
	c := &buffalo.DefaultContext{
		Context: context.Background(),
	}
	sp := buffet.SpanFromContext(c)
	sp.SetOperationName("test.storage.mongo.List")
	defer sp.Finish()

	r := m.Require()
	versions := []string{"v1.0.0", "v1.1.0", "v1.2.0"}
	for _, version := range versions {
		m.storage.Save(c, module, version, mod, zip, info)
	}
	retVersions, err := m.storage.List(module)
	r.NoError(err)
	r.Equal(versions, retVersions)
}
