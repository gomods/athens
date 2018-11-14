package download

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
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

	// listHandler := LogEntryHandler(ListHandler, opts)
	// app.GET(PathList, noCacheMw(listHandler))

	latestHandler := LogEntryHandler(LatestHandler, opts)
	app.GET(PathLatest, noCacheMw(latestHandler))

	app.GET(PathVersionInfo, LogEntryHandler(VersionInfoHandler, opts))
	app.GET(PathVersionModule, LogEntryHandler(VersionModuleHandler, opts))
	app.GET(PathVersionZip, LogEntryHandler(VersionZipHandler, opts))
}

// ProtocolHandlerV2 is the buffalo-less implementation of ProtocolHandler
type ProtocolHandlerV2 func(dp Protocol, lggr log.Entry, eng *render.Engine) http.HandlerFunc

// LogEntryHandlerV2 mimics the behavior of LogEntryHandler, but for a basic http.Handler
func LogEntryHandlerV2(ph ProtocolHandlerV2, opts *HandlerOpts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ent := opts.Logger.WithFields(logrus.Fields{
			"http-method": r.Method,
			"http-path":   r.URL.Path,
			"http-url":    r.URL.String(),
		})
		handler := ph(opts.Protocol, ent, opts.Engine)
		handler(w, r)
	}
}

// RegisterHandlersV2 is a transition function that will allow us to slowly define routes
// on buffalo's internal *mux.Router. As route handlers are migrated from buffalo.Handlers
// to regular http.Handlers the route definition can be moved from the existing RegisterHandlers
// to this method. When finally removing buffalo, this method can be renamed to RegisterHandlers
// and a new mux.Router can be passed to this function instead of the underlying buffalo Muxer
func RegisterHandlersV2(m *mux.Router, opts *HandlerOpts) {
	// If true, this would only panic at boot time, static nil checks anyone?
	if opts == nil || opts.Protocol == nil || opts.Engine == nil || opts.Logger == nil {
		panic("absolutely unacceptable handler opts")
	}

	noCacheMw := middleware.CacheControlV2("no-cache, no-store, must-revalidate")

	listHandler := LogEntryHandlerV2(ListHandlerBasic, opts)
	// because of some weirdness in how buffalo handles slashes, a trailing slash will have to be appended
	// to every route definition until moving off of buffalo. https://gobuffalo.io/en/docs/routing#loose-slash
	m.Handle(PathList+"/", noCacheMw(http.HandlerFunc(listHandler))).Methods("GET")
}
