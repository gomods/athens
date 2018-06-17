package gcp

import (
	"github.com/gobuffalo/buffalo"
	"google.golang.org/appengine/aetest"
)

func (g *GcpTests) TestNewStorage() {
	r := g.Require()
	ctx, done, err := aetest.NewContext()
	defer done()
	r.NoError(err)
	c := &buffalo.DefaultContext{}
	c.Context = ctx
	store, err := New(c, g.options)
	r.NoError(err)
	r.NotNil(store.bucket)
}
