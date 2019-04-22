package download

import "net/http"

const (
	AsyncFetchWebhookPath     = "/download/async"
	SyncAliasFetchWebhookPath = "/alias/sync"
)

func AsyncFetchWebhookHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// TODO
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("WIP"))
	})
}

func SyncAliasFetchWebhookHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// TODO
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("WIP"))
	})
}
