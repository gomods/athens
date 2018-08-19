package actions

import (
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/module"
	"github.com/markbates/willie"
	"github.com/stretchr/testify/require"
)

func middlewareApp(filterFile, olympusEndpoint string) (*buffalo.App, error) {
	h := func(c buffalo.Context) error {
		return c.Render(200, nil)
	}

	a := buffalo.New(buffalo.Options{})
	mf, err := newTestFilter(filterFile)
	if err != nil {
		return nil, err
	}
	a.Use(newFilterMiddleware(mf, olympusEndpoint))
	initializeTracing(a)

	a.GET(download.PathList, h)
	return a, nil
}

func newTestFilter(filterFile string) (*module.Filter, error) {

	f, err := module.NewFilter(filterFile)
	if err != nil {
		return nil, err
	}
	f.AddRule("github.com/gomods/athens/", module.Include)
	f.AddRule("github.com/athens-artifacts/no-tags", module.Exclude)
	f.AddRule("github.com/athens-artifacts", module.Direct)
	return f, nil
}

func Test_Middleware(t *testing.T) {
	r := require.New(t)
	conf, err := getConf()
	if err != nil {
		t.Error(err)
	}

	filterFile := conf.FilterFile
	olympusEndpoint := conf.Proxy.OlympusGlobalEndpoint

	mw, err := middlewareApp(filterFile, olympusEndpoint)
	r.NoError(err)
	w := willie.New(mw)

	// Public, expects to be redirected to olympus
	res := w.Request("/github.com/gomods/athens/@v/list").Get()
	r.Equal(303, res.Code)
	r.Equal(olympusEndpoint+"/github.com/gomods/athens/@v/list", res.HeaderMap.Get("Location"))

	// Excluded, expects a 403
	res = w.Request("/github.com/athens-artifacts/no-tags/@v/list").Get()
	r.Equal(403, res.Code)

	// Private, the proxy is working and returns a 200
	res = w.Request("/github.com/athens-artifacts/happy-path/@v/list").Get()
	r.Equal(200, res.Code)
}
