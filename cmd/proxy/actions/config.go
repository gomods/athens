package actions

import (
	"encoding/json"
	"net/http"

	"github.com/gomods/athens/pkg/config"
)

// ConfigHandler returns the current configuration as json when hitting the /config url
func ConfigHandler(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config" {
			w.WriteHeader(http.StatusForbidden)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(c)
	}
}
