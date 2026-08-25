package download

import (
	"net/http"
	"net/url"
	"path"

	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/middleware"
	"github.com/gorilla/mux"
)

// ProtocolHandler is a function that takes all that it needs to return
// a ready-to-go http handler that serves up cmd/go's download protocol.
type ProtocolHandler func(dp Protocol, lggr log.Entry, df *mode.DownloadFile) http.Handler

// HandlerOpts are the generic options for a ProtocolHandler.
type HandlerOpts struct {
	Protocol     Protocol
	Logger       *log.Logger
	DownloadFile *mode.DownloadFile
	// ContentCacheControl is the Cache-Control header value sent on the info,
	// mod and zip endpoints. Those endpoints serve immutable module content, so
	// a caching proxy in front of Athens can safely cache them. When empty (the
	// default) no Cache-Control header is set on them.
	ContentCacheControl string
}

// LogEntryHandler pulls a log entry from the request context. Thanks to the
// LogEntryMiddleware, we should have a log entry stored in the context for each
// request with request-specific fields. This will grab the entry and pass it to
// the protocol handlers.
func LogEntryHandler(ph ProtocolHandler, opts *HandlerOpts) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		ent := log.EntryFromContext(r.Context())
		handler := ph(opts.Protocol, ent, opts.DownloadFile)
		handler.ServeHTTP(w, r)
	}
	return http.HandlerFunc(f)
}

// RegisterHandlers is a convenience method that registers
// all the download protocol paths for you.
func RegisterHandlers(r *mux.Router, opts *HandlerOpts) {
	// If true, this would only panic at boot time, static nil checks anyone?
	if opts == nil || opts.Protocol == nil || opts.Logger == nil {
		panic("absolutely unacceptable handler opts")
	}
	noCacheMw := middleware.CacheControl("no-cache, no-store, must-revalidate")

	listHandler := LogEntryHandler(ListHandler, opts)
	r.Handle(PathList, noCacheMw(listHandler))

	latestHandler := LogEntryHandler(LatestHandler, opts)
	r.Handle(PathLatest, noCacheMw(latestHandler)).Methods(http.MethodGet)

	infoHandler := LogEntryHandler(InfoHandler, opts)
	moduleHandler := LogEntryHandler(ModuleHandler, opts)
	zipHandler := LogEntryHandler(ZipHandler, opts)
	if opts.ContentCacheControl != "" {
		cacheMw := middleware.CacheControl(opts.ContentCacheControl)
		infoHandler = cacheMw(infoHandler)
		moduleHandler = cacheMw(moduleHandler)
		zipHandler = cacheMw(zipHandler)
	}

	r.Handle(PathVersionInfo, infoHandler).Methods(http.MethodGet)
	r.Handle(PathVersionModule, moduleHandler).Methods(http.MethodGet)
	r.Handle(PathVersionZip, zipHandler).Methods(http.MethodGet, http.MethodHead)
}

func getRedirectURL(base, downloadPath string) (string, error) {
	url, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	url.Path = path.Join(url.Path, downloadPath)
	return url.String(), nil
}
