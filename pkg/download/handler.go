package download

import (
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/middleware"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gorilla/mux"
	"golang.org/x/mod/semver"
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
		fileMw = moduleFileCacheControl(opts.CacheControl)
	}

	r.Handle(PathVersionInfo, fileMw(LogEntryHandler(InfoHandler, opts))).Methods(http.MethodGet)
	r.Handle(PathVersionModule, fileMw(LogEntryHandler(ModuleHandler, opts))).Methods(http.MethodGet)
	r.Handle(PathVersionZip, fileMw(LogEntryHandler(ZipHandler, opts))).Methods(http.MethodGet, http.MethodHead)
}

// getRedirectURL returns the URL at base that serves the module file requested by r.
func getRedirectURL(base string, r *http.Request) (string, error) {
	url, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	url.Path = path.Join(url.Path, paths.ProtocolPath(r))
	return url.String(), nil
}

// moduleFileCacheControl returns middleware that sets the Cache-Control header for module file endpoints
// (.info, .mod, and .zip). The configured CacheControl header is only applied to HTTP 200 responses
// for canonical semantic versions. Dynamic branch/revision queries (e.g. /@v/main.info) receive
// "no-cache, no-store, must-revalidate", and non-200 responses (such as 404 or errors) never receive
// an immutable or long cache header.
func moduleFileCacheControl(cacheControl string) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {
			ver, err := paths.GetVersion(r)
			isCanonical := err == nil && semver.IsValid(ver) && semver.Canonical(ver) == ver
			rw := &moduleFileResponseWriter{
				ResponseWriter: w,
				cacheControl:   cacheControl,
				isCanonical:    isCanonical,
				wroteHeader:    false,
			}
			h.ServeHTTP(rw, r)
			if !rw.wroteHeader {
				rw.WriteHeader(http.StatusOK)
			}
		}
		return http.HandlerFunc(f)
	}
}

type moduleFileResponseWriter struct {
	http.ResponseWriter

	cacheControl string
	isCanonical  bool
	wroteHeader  bool
}

func (w *moduleFileResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	if statusCode == http.StatusOK {
		if w.isCanonical && w.cacheControl != "" {
			w.Header().Set("Cache-Control", w.cacheControl)
		} else {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}
	} else {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *moduleFileResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func (w *moduleFileResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *moduleFileResponseWriter) ReadFrom(src io.Reader) (int64, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(src)
	}
	return io.Copy(w.ResponseWriter, src)
}
