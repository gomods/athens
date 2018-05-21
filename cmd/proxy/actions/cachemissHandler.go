package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
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
			fmt.Println("2")
			if err != nil {
				fmt.Println(err)
				return nextErr
			}

			cm := payloads.Module{Name: params.module, Version: params.version}
			content, err := json.Marshal(cm)
			fmt.Println("3")
			if err != nil {
				fmt.Println(err)
				return nextErr
			}

			olympusEndpoint := getCurrentOlympus()
			fmt.Println("4")
			if olympusEndpoint == "" || olympusEndpoint == OlympusGlobalEndpoint {
				fmt.Println(err)
				return nextErr
			}

			req, err := http.NewRequest("POST", olympusEndpoint+"/cachemiss", bytes.NewBuffer(content))

			fmt.Println("5")
			fmt.Println(olympusEndpoint)
			if err != nil {
				fmt.Println(err)
				return nextErr
			}

			fmt.Println("6")
			client := http.Client{Timeout: 30 * time.Second}
			fmt.Println(client.Do(req))
		}

		return nextErr
	}
}

func isModuleNotFoundErr(err error) bool {
	s := err.Error()
	return strings.HasPrefix(s, "module ") && strings.HasSuffix(s, "not found")
}
