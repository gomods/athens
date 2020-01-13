package middleware

import (
	"bytes"
	"context"
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
				response, err := validate(validatorHook, mod, version)
				if err != nil {
					entry := log.EntryFromContext(r.Context())
					entry.SystemErr(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				maybeLogValidationReason(r.Context(), string(response.Message), mod, version)

				if !response.Valid {
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}
			h.ServeHTTP(w, r)
		})
	}
}

func maybeLogValidationReason(context context.Context, message string, mod string, version string) {
	if len(message) > 0 {
		entry := log.EntryFromContext(context)
		entry.Warnf("error validating %s@%s %s", mod, version, message)
	}
}

type validationParams struct {
	Module  string
	Version string
}

type ValidationResponse struct {
	Valid   bool
	Message []byte
}

func validate(hook, mod, ver string) (ValidationResponse, error) {
	const op errors.Op = "actions.validate"

	toVal := &validationParams{mod, ver}
	jsonVal, err := json.Marshal(toVal)
	if err != nil {
		return ValidationResponse{Valid: false}, errors.E(op, err)
	}

	resp, err := http.Post(hook, "application/json", bytes.NewBuffer(jsonVal))
	if err != nil {
		return ValidationResponse{Valid: false}, errors.E(op, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return maybeReadResponseReason(resp), nil
	case http.StatusForbidden:
		return maybeReadResponseReason(resp), nil
	default:
		return ValidationResponse{Valid: false}, errors.E(op, "Unexpected status code ", resp.StatusCode)
	}
}

func maybeReadResponseReason(resp *http.Response) ValidationResponse {
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return ValidationResponse{Valid: resp.StatusCode == http.StatusOK, Message: body}
}
