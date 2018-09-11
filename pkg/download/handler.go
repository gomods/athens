package download

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/middleware"
)

// ProtocolHandler is a function that takes all that it needs to return
// a ready-to-go buffalo handler that serves up cmd/go's download protocol.
type ProtocolHandler func(dp Protocol, lggr log.Entry, eng *render.Engine) buffalo.Handler

// HandlerOpts are the generic options
// for a ProtocolHandler
type HandlerOpts struct {
	Protocol Protocol
	Logger   *log.Logger
	Engine   *render.Engine
}

// RegisterHandlers is a convenience method that registers
// all the download protocol paths for you.
func RegisterHandlers(app *buffalo.App, opts *HandlerOpts) {
	// If true, this would only panic at boot time, static nil checks anyone?
	if opts == nil || opts.Protocol == nil || opts.Engine == nil || opts.Logger == nil {
		panic("absolutely unacceptable handler opts")
	}
	noCacheMw := middleware.CacheControl("no-cache, no-store, must-revalidate")

	listHandler := ListHandler(opts.Protocol, opts.Engine)
	app.GET(PathList, noCacheMw(listHandler))

	latestHandler := LatestHandler(opts.Protocol, opts.Engine)
	app.GET(PathLatest, noCacheMw(latestHandler))

	app.GET(PathVersionInfo, VersionInfoHandler(opts.Protocol, opts.Engine))
	app.GET(PathVersionModule, VersionModuleHandler(opts.Protocol, opts.Engine))
	app.GET(PathVersionZip, VersionZipHandler(opts.Protocol, opts.Engine))
}
