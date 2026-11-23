package actions

import (
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// newHTTPClients builds the outbound HTTP clients used by the proxy.
//
// The validator client is used for the validation webhook, whose bodies are
// small JSON documents, so a total Client.Timeout is appropriate: a hung
// webhook must not block every versioned module request.
//
// The storage client is shared by the S3 and external storage backends, which
// stream large module zips. Client.Timeout also bounds reading the response
// body and would cut those transfers off, so it is left unset. Instead the
// transport's ResponseHeaderTimeout bounds only the wait for the first
// response byte, which protects against a hung server without limiting the
// body stream.
func newHTTPClients(timeout time.Duration) (validator, storage *http.Client) {
	validator = &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   timeout,
	}
	storage = &http.Client{
		Transport: otelhttp.NewTransport(newStorageTransport(timeout)),
	}
	return validator, storage
}

// newStorageTransport returns a clone of http.DefaultTransport with
// ResponseHeaderTimeout set to timeout (when positive). It never mutates the
// shared default transport.
func newStorageTransport(timeout time.Duration) *http.Transport {
	var tr *http.Transport
	if dt, ok := http.DefaultTransport.(*http.Transport); ok {
		tr = dt.Clone()
	} else {
		tr = &http.Transport{Proxy: http.ProxyFromEnvironment}
	}
	if timeout > 0 {
		tr.ResponseHeaderTimeout = timeout
	}
	return tr
}
