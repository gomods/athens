package actions

import (
	"net/http"
)

func proxyHomeHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(`"Welcome to The Athens Proxy"`))
}
