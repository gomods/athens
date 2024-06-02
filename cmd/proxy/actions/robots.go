package actions

import (
	"net/http"

	"github.com/gomods/athens/pkg/config"
)

// robotsHandler implements GET baseURL/robots.txt.
func robotsHandler(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.ServeFile(w, r, c.RobotsFile)
	}
}
