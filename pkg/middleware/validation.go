package middleware

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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
				valid, err := validate(validatorHook, mod, version)
				if err != nil {
					entry := log.EntryFromContext(r.Context())
					entry.SystemErr(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if !valid.Success {
					entry := log.EntryFromContext(r.Context())
					entry.Warnf("Module %s:%s %s", mod, version, valid.Reason)
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

type ValidationResponse struct {
	Success bool
	Reason  string
}

func validate(hook, mod, ver string) (ValidationResponse, error) {
	const op errors.Op = "actions.validate"

	toVal := &validationParams{mod, ver}
	jsonVal, err := json.Marshal(toVal)
	if err != nil {
		return ValidationResponse{Success: false}, errors.E(op, err)
	}

	resp, err := http.Post(hook, "application/json", bytes.NewBuffer(jsonVal))
	if err != nil {
		return ValidationResponse{Success: false}, errors.E(op, err)
	}

	switch {
	case resp.StatusCode == http.StatusOK:
		return ValidationResponse{Success: true}, nil
	case resp.StatusCode == http.StatusForbidden:
		return ValidationResponse{Success: false, Reason: maybeReadResponseReason(resp)}, nil
	default:
		return ValidationResponse{Success: false}, errors.E(op, "Unexpected status code ", resp.StatusCode)
	}
}

func maybeReadResponseReason(resp *http.Response) string {
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if len(bodyBytes) > 0 {
		return string(bodyBytes)
	}

	return "unknown"
}
