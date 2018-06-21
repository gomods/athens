package gcp

import (
	"google.golang.org/appengine/aetest"
)

func (g *GcpTests) TestNewStorage() {
	r := g.Require()
	ctx, done, err := aetest.NewContext()
	defer done()
	r.NoError(err)
	store, err := New(ctx, "staging.praxis-cab-207400.appspot.com", g.options)
	r.NoError(err)
	r.NotNil(store.bucket)
}

func (g *GcpTests) TestSave() {
	r := g.Require()
	ctx, done, err := aetest.NewContext()
	defer done()
	r.NoError(err)
	store, err := New(ctx, "staging.praxis-cab-207400.appspot.com", g.options)
	r.NoError(err)
	err = store.Save(ctx, "testface", "v1.2.3", mod, info, zip)
	r.NoError(err)
}
