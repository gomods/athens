package actions

import (
	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/download/addons"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/stash"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gorilla/mux"
	"github.com/spf13/afero"
)

func addProxyRoutes(
	r *mux.Router,
	s storage.Backend,
	l *log.Logger,
	goBin string,
	goGetWorkers,
	protocolWorkers int,
) error {
	r.HandleFunc("/", proxyHomeHandler)
	r.HandleFunc("/healthz", healthHandler)
	r.HandleFunc("/readyz", getReadinessHandler(s))
	r.HandleFunc("/version", versionHandler)

	// Download Protocol
	// the download.Protocol and the stash.Stasher interfaces are composable
	// in a middleware fashion. Therefore you can separate concerns
	// by the functionality: a download.Protocol that just takes care
	// of "go getting" things, and another Protocol that just takes care
	// of "pooling" requests etc.

	// In our case, we'd like to compose both interfaces in a particular
	// order to ensure logical ordering of execution.

	// Here's the order of an incoming request to the download.Protocol:

	// 1. The downloadpool gets hit first, and manages concurrent requests
	// 2. The downloadpool passes the request to its parent Protocol: stasher
	// 3. The stasher Protocol checks storage first, and if storage is empty
	// it makes a Stash request to the stash.Stasher interface.

	// Once the stasher picks up an order, here's how the requests go in order:
	// 1. The singleflight picks up the first request and latches duplicate ones.
	// 2. The singleflight passes the stash to its parent: stashpool.
	// 3. The stashpool manages limiting concurrent requests and passes them to stash.
	// 4. The plain stash.New just takes a request from upstream and saves it into storage.
	fs := afero.NewOsFs()
	mf, err := module.NewGoGetFetcher(goBin, fs)
	if err != nil {
		return err
	}

	lister := download.NewVCSLister(goBin, fs)
	st := stash.New(mf, s, stash.WithPool(goGetWorkers), stash.WithSingleflight)

	dpOpts := &download.Opts{
		Storage: s,
		Stasher: st,
		Lister:  lister,
	}
	dp := download.New(dpOpts, addons.WithPool(protocolWorkers))

	handlerOpts := &download.HandlerOpts{Protocol: dp, Logger: l}
	download.RegisterHandlers(r, handlerOpts)

	storage.RegisterHandlers(r, s)

	return nil
}
