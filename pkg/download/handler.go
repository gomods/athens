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

// LogEntryHandler pulls a log entry from the buffalo context. Thanks to the
// LogEntryMiddleware, we should have a log entry stored in the context for each
// request with request-specific fields. This will grab the entry and pass it to
// the protocol handlers
func LogEntryHandler(ph ProtocolHandler, opts *HandlerOpts) buffalo.Handler {
	return func(c buffalo.Context) error {
		ent := log.EntryFromContext(c)
		handler := ph(opts.Protocol, ent, opts.Engine)

		return handler(c)
	}
}

// RegisterHandlers is a convenience method that registers
// all the download protocol paths for you.
func RegisterHandlers(app *buffalo.App, opts *HandlerOpts) {
	// If true, this would only panic at boot time, static nil checks anyone?
	if opts == nil || opts.Protocol == nil || opts.Engine == nil || opts.Logger == nil {
		panic("absolutely unacceptable handler opts")
	}
	noCacheMw := middleware.CacheControl("no-cache, no-store, must-revalidate")

	listHandler := LogEntryHandler(ListHandler, opts)
	app.GET(PathList, noCacheMw(listHandler))

	latestHandler := LogEntryHandler(LatestHandler, opts)
	app.GET(PathLatest, noCacheMw(latestHandler))

	app.GET(PathVersionInfo, LogEntryHandler(VersionInfoHandler, opts))
	app.GET(PathVersionModule, LogEntryHandler(VersionModuleHandler, opts))
	app.GET(PathVersionZip, LogEntryHandler(VersionZipHandler, opts))
	app.GET(PathCatalog, LogEntryHandler(CatalogHandler, opts))

}
