package actions

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gomods/athens/pkg/payloads"

	"github.com/gobuffalo/buffalo"
)

func cachemissHandler(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		nextErr := next(c)

		if httperr, ok := nextErr.(buffalo.HTTPError); ok && httperr.Status == http.StatusNotFound {
			// TODO: set workers and process it there to minimize latency
			params, err := getAllPathParams(c)
			if err != nil {
				return nextErr
			}

			cm := payloads.Module{Name: params.module, Version: params.version}
			content, err := json.Marshal(cm)
			if err != nil {
				return nextErr
			}

			olympusEndpoint := getCurrentOlympus()
			if olympusEndpoint == "" || olympusEndpoint == OlympusGlobalEndpoint {
				return nextErr
			}

			req, err := http.NewRequest("POST", olympusEndpoint+"/cachemiss", bytes.NewBuffer(content))
			if err != nil {
				return nextErr
			}

			client := http.Client{Timeout: 30 * time.Second}
			client.Do(req)
		}

		return nextErr
	}
}
