package actions

import (
	"net/http"

	"github.com/gomods/athens/pkg/config"
)

// robotsHandler implements GET baseURL/robots.txt
func robotsHandler(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, c.RobotsFile)
	}
}
