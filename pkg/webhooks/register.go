package webhooks

import (
	"net/http"

	"github.com/gomods/athens/pkg/download"
	"github.com/gorilla/mux"
)

func RegisterHandlers(r *mux.Router, opts *download.HandlerOpts) {

	// TODO: this webhook simply queues up a fetch opteration in the
	// background and immediately returns a 201 CREATED response, without
	// returning a body
	r.Handle(
		asyncFetchPath,
		download.LogEntryHandler(asyncFetchHandler, opts),
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
		syncAliasFetchPath,
		download.LogEntryHandler(syncAliasFetchHandler, opts),
	).Methods(http.MethodPost)
}
