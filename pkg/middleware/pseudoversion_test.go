package middleware

import (
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gomods/athens/pkg/log"
	"github.com/markbates/willie"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func middlewarePseudoverApp(fs afero.Fs) *buffalo.App {
	h := func(c buffalo.Context) error {
		return c.Render(200, nil)
	}

	a := buffalo.New(buffalo.Options{})
	goBinaryPath := envy.Get("GO_BINARY_PATH", "go")
	lggr := log.New("none", logrus.DebugLevel)
	a.Use(NewPseudoversionMiddleware(lggr, fs, goBinaryPath))
	a.GET(pathList, h)
	a.GET(pathVersionInfo, h)
	return a
}

func Test_FilterPseudoversion(t *testing.T) {
	r := require.New(t)
	fs := afero.NewOsFs()

	app := middlewarePseudoverApp(fs)
	w := willie.New(app)

	// List, no change
	res := w.Request("/github.com/gomods/athens/@v/list").Get()
	r.Equal(200, res.Code)

	// Hash, expects redirect to pseudover
	res = w.Request("/github.com/arschles/assert/@v/fc2da9844984ce5093111298174706e14d4c0c47.info").Get()
	r.Equal(303, res.Code)
	r.Equal("/github.com/arschles/assert/@v/v0.0.0-20160620175154-fc2da9844984.info", res.HeaderMap.Get("Location"))

	// Normal version, no redirect
	res = w.Request("/github.com/arschles/assert/@v/v1.0.0.info").Get()
	r.Equal(200, res.Code)
}
