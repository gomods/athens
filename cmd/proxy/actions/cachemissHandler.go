package actions

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gomods/athens/pkg/payloads"

	"github.com/gobuffalo/buffalo"
)

func cachemissHandler(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		nextErr := next(c)

		if isModuleNotFoundErr(nextErr) {
			// TODO: set workers and process it there to minimize latency
			params, err := getAllPathParams(c)
			if err != nil {
				log.Println(err)
				return nextErr
			}

			cm := payloads.Module{Name: params.module, Version: params.version}
			content, err := json.Marshal(cm)
			if err != nil {
				log.Println(err)
				return nextErr
			}

			olympusEndpoint := getCurrentOlympus()
			if olympusEndpoint == "" || olympusEndpoint == OlympusGlobalEndpoint {
				return nextErr
			}

			req, err := http.NewRequest("POST", olympusEndpoint+"/cachemiss", bytes.NewBuffer(content))
			if err != nil {
				log.Println(err)
				return nextErr
			}

			client := http.Client{Timeout: 30 * time.Second}
			client.Do(req)
		}

		return nextErr
	}
}

func isModuleNotFoundErr(err error) bool {
	s := err.Error()
	return strings.HasPrefix(s, "module ") && strings.HasSuffix(s, "not found")
}
