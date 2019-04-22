package download

import (
	"net/http"

	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/middleware"
	"github.com/gorilla/mux"
)

// ProtocolHandler is a function that takes all that it needs to return
// a ready-to-go http handler that serves up cmd/go's download protocol.
type ProtocolHandler func(dp Protocol, lggr log.Entry) http.Handler

// HandlerOpts are the generic options
// for a ProtocolHandler
type HandlerOpts struct {
	Protocol Protocol
	Logger   *log.Logger
}

// LogEntryHandler pulls a log entry from the request context. Thanks to the
// LogEntryMiddleware, we should have a log entry stored in the context for each
// request with request-specific fields. This will grab the entry and pass it to
// the protocol handlers
func LogEntryHandler(ph ProtocolHandler, opts *HandlerOpts) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		ent := log.EntryFromContext(r.Context())
		handler := ph(opts.Protocol, ent)
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

	r.Handle(PathVersionInfo, LogEntryHandler(InfoHandler, opts)).Methods(http.MethodGet)
	r.Handle(PathVersionModule, LogEntryHandler(ModuleHandler, opts)).Methods(http.MethodGet)
	r.Handle(PathVersionZip, LogEntryHandler(ZipHandler, opts)).Methods(http.MethodGet)
}

func RegisterWebhookHandlers(r *mux.Router, opts *HandlerOpts) {

	// TODO: this webhook simply queues up a fetch opteration in the
	// background and immediately returns a 201 CREATED response, without
	// returning a body
	r.Handle(
		AsyncFetchWebhookPath,
		LogEntryHandler(AsyncFetchWebhookHandler, opts),
	).Methods(http.MethodPost)
	// TODO: this webhook does a synchronous fetch for a VCS path and stores
	// it locally under an alias. The endpoint will return 200 after
	// the code for the version listed under the /@latest endpoint is fetched
	// and stored.
	//
	// This endpoint is in place to allow Athens to serve vanity import paths
	// whose code is backed by a VCS.
	//
	// Future requests to the same vanity import path will cause a
	// fetch to the same backing VCS. This endpoint will not work
	// for future vanity import paths that have the original path as their
	// prefix. Concretely:
	//
	//	- Vanity path vanity.dev is registered with this endpoint as vcs.dev
	//	- A go get vanity.dev@v1.0.2 happens and the code comes from vcs.dev
	//	- A go get vanity.dev/v2 happens
	//
	// In this case, the code for vanity.dev/v2 will not be fetches from vcs.dev
	r.Handle(
		SyncAliasFetchWebhookPath,
		LogEntryHandler(SyncAliasFetchWebhookHandler, opts),
	).Methods(http.MethodPost)
}
