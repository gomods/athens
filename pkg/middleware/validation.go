package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gorilla/mux"
)

// NewValidationMiddleware builds a middleware function that performs validation checks by calling
// an external webhook.
func NewValidationMiddleware(client *http.Client, validatorHook string) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mod, err := paths.GetModule(r)
			if err != nil {
				// if there is no module the path we are hitting is not one related to modules, like /
				h.ServeHTTP(w, r)
				return
			}
			ctx := r.Context()
			// not checking the error. Not all requests include a version
			// i.e. list requests path is like /{module:.+}/@v/list with no version parameter
			version, _ := paths.GetVersion(r)
			if version != "" {
				response, err := validate(ctx, client, validatorHook, mod, version)
				if err != nil {
					entry := log.EntryFromContext(ctx)
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

func maybeLogValidationReason(context context.Context, message, mod, version string) {
	if len(message) > 0 {
		entry := log.EntryFromContext(context)
		entry.Warnf("error validating %s@%s %s", mod, version, message)
	}
}

type validationParams struct {
	Module  string
	Version string
}

type validationResponse struct {
	Valid   bool
	Message []byte
}

func validate(ctx context.Context, client *http.Client, hook, mod, ver string) (validationResponse, error) {
	const op errors.Op = "actions.validate"

	toVal := &validationParams{mod, ver}
	jsonVal, err := json.Marshal(toVal)
	if err != nil {
		return validationResponse{Valid: false}, errors.E(op, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hook, bytes.NewReader(jsonVal))
	if err != nil {
		return validationResponse{}, errors.E(op, err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return validationResponse{Valid: false}, errors.E(op, err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		return validationResponseFromRequest(resp), nil
	case http.StatusForbidden:
		return validationResponseFromRequest(resp), nil
	default:
		return validationResponse{Valid: false}, errors.E(op, "Unexpected status code ", resp.StatusCode)
	}
}

func validationResponseFromRequest(resp *http.Response) validationResponse {
	body, _ := io.ReadAll(resp.Body)
	return validationResponse{Valid: resp.StatusCode == http.StatusOK, Message: body}
}
