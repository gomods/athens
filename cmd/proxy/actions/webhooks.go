package actions

import (
	"encoding/json"
	"net/http"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gorilla/mux"
)

func addWebhooks(
	r *mux.Router,
	s storage.Backend,
	l *log.Logger,
	c *config.Config,
) error {
	r.HandleFunc("/webhooks/bulk", bulkAsyncWebhookHandler(l, s)).Methods("POST")

	return nil
}

func bulkAsyncWebhookHandler(l *log.Logger, s storage.Backend) func(http.ResponseWriter, *http.Request) {
	type webhookModule struct {
		ModuleName     string   `json:"module_name"`
		ModuleVersions []string `json:"module_versions"`
	}

	type webhooksPostBody struct {
		ModulesToFetch []webhookModule `json:"modules_to_fetch"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &webhooksPostBody{}
		defer r.Body.Close()
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(req); err != nil {
			// TODO: return error to HTTP client
			return
		}

		for _, moduleToFetch := range req.ModulesToFetch {
			for _, _ /*moduleVersion*/ := range moduleToFetch.ModuleVersions {

				// TODO: do a get/fetch/store (using a stasher?) for
				// (moduleToFetch.ModuleName, moduleVersion)
			}
		}
	}
}
