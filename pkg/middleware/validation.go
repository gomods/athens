package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gorilla/mux"
)

// NewValidationMiddleware builds a middleware function that performs validation checks by calling
// an external webhook
func NewValidationMiddleware(validatorHook string) mux.MiddlewareFunc {
	const op errors.Op = "actions.NewValidationMiddleware"
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mod, err := paths.GetModule(r)
			if err != nil {
				// if there is no module the path we are hitting is not one related to modules, like /
				h.ServeHTTP(w, r)
				return
			}
			// not checking the error. Not all requests include a version
			// i.e. list requests path is like /{module:.+}/@v/list with no version parameter
			version, _ := paths.GetVersion(r)
			if version != "" {
				valid, err := validate(validatorHook, mod, version, c.Request().Header.Get("Authorization"))
				if err != nil {
					entry := log.EntryFromContext(r.Context())
					entry.SystemErr(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if !valid {
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}
			h.ServeHTTP(w, r)
		})
	}
}

type validationParams struct {
	Module  string
	Version string
}

func validate(hook, mod, ver, auth string) (bool, error) {
	const op errors.Op = "actions.validate"

	toVal := &validationParams{mod, ver}
	jsonVal, err := json.Marshal(toVal)
	if err != nil {
		return false, errors.E(op, err)
	}

	req, err := http.NewRequest(http.MethodPost, hook, bytes.NewBuffer(jsonVal))
	if err != nil {
		return false, errors.E(op, err)
	}

	req.Header.Add("Content-Type", "application/json")
	if auth != "" {
		req.Header.Add("Authorization", auth)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, errors.E(op, err)
	}

	switch {
	case resp.StatusCode == http.StatusOK:
		return true, nil
	case resp.StatusCode == http.StatusForbidden:
		return false, nil
	default:
		return false, errors.E(op, "Unexpected status code ", resp.StatusCode)
	}
}
