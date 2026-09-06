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
	// CacheControl, when non-empty, is sent as the Cache-Control header on the
	// module file endpoints (.info, .mod and .zip). Those responses are immutable
	// for a given module@version, so a value like "public, max-age=..." lets a
	// caching proxy in front of Athens serve them without hitting Athens again.
	CacheControl string
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

	// The .info, .mod and .zip files are immutable for a given module@version,
	// so honor an operator supplied Cache-Control header on them if one is set.
	fileMw := func(h http.Handler) http.Handler { return h }
	if opts.CacheControl != "" {
		fileMw = middleware.CacheControl(opts.CacheControl)
	}

	r.Handle(PathVersionInfo, fileMw(LogEntryHandler(InfoHandler, opts))).Methods(http.MethodGet)
	r.Handle(PathVersionModule, fileMw(LogEntryHandler(ModuleHandler, opts))).Methods(http.MethodGet)
	r.Handle(PathVersionZip, fileMw(LogEntryHandler(ZipHandler, opts))).Methods(http.MethodGet, http.MethodHead)
}

func getRedirectURL(base, downloadPath string) (string, error) {
	url, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	url.Path = path.Join(url.Path, downloadPath)
	return url.String(), nil
}
