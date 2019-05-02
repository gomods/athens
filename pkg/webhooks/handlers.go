package webhooks

import (
	"net/http"

	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/log"
)

const (
	asyncFetchPath     = "/download/async"
	syncAliasFetchPath = "/alias/sync"
)

// the two anonymous parameters are needed so that we can pass this
// handler to download.LogEntryHandler
func asyncFetchHandler(download.Protocol, log.Entry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// TODO
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("WIP"))
	})
}

// the two anonymous parameters are needed so that we can pass this
// handler to download.LogEntryHandler
func syncAliasFetchHandler(download.Protocol, log.Entry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// TODO
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("WIP"))
	})
}
