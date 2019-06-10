package actions

import (
	"net/http"
)

// robotsHandler implements GET /robots.txt
func robotsHandler(s string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(s))
	}
}
