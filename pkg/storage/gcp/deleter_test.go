package gcp

import (
	"bytes"
	"time"
)

func (g *GcpTests) TestDeleter() {
	r := g.Require()
	store, err := NewWithCredentials(g.context, g.options)
	r.NoError(err)

	version := "delete" + time.Now().String()
	err = store.Save(g.context, g.module, version, mod, bytes.NewReader(zip), info)
	r.NoError(err)

	err = store.Delete(g.module, version)
	r.NoError(err)

	exists := store.Exists(g.module, version)
	r.Equal(false, exists)
}
