package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/bketelsen/buffet"
	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/config/env"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/paths"
	"github.com/sirupsen/logrus"
)

type middlewareFunc func(entry log.Entry) buffalo.MiddlewareFunc

// LogEntryMiddleware builds a log.Entry applying the request parameter to the given
// log.Logger and propagates it to the given MiddlewareFunc
func LogEntryMiddleware(middleware middlewareFunc, lggr *log.Logger) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			req := c.Request()
			ent := lggr.WithFields(logrus.Fields{
				"http-method": req.Method,
				"http-path":   req.URL.Path,
				"http-url":    req.URL.String(),
			})
			m := middleware(ent)
			return m(next)(c)
		}
	}
}

// NewFilterMiddleware builds a middleware function that implements the filters configured in
// the filter file.
func NewFilterMiddleware(mf *module.Filter) buffalo.MiddlewareFunc {
	const op errors.Op = "actions.FilterMiddleware"

	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			sp := buffet.SpanFromContext(c).SetOperationName("filterMiddleware")
			defer sp.Finish()

			mod, err := paths.GetModule(c)

			if err != nil {
				// if there is no module the path we are hitting is not one related to modules, like /
				return next(c)
			}

			// not checking the error. Not all requests include a version
			// i.e. list requests path is like /{module:.+}/@v/list with no version parameter
			version, _ := paths.GetVersion(c)

			if isPseudoVersion(version) {
				return next(c)
			}

			rule := mf.Rule(mod)
			switch rule {
			case module.Exclude:
				return c.Render(http.StatusForbidden, nil)
			case module.Direct:
				return next(c)
			case module.Include:
				// TODO : spin up cache filling worker and serve the request using the cache
				newURL := redirectToOlympusURL(c.Request().URL)
				return c.Redirect(http.StatusSeeOther, newURL)
			}

			return next(c)
		}
	}
}

// NewValidationMiddleware builds a middleware function that performs validation checks by calling
// an external webhook
func NewValidationMiddleware(entry log.Entry) buffalo.MiddlewareFunc {
	const op errors.Op = "actions.ValidationMiddleware"

	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			sp := buffet.SpanFromContext(c).SetOperationName("validationMiddleware")
			defer sp.Finish()

			mod, err := paths.GetModule(c)

			if err != nil {
				// if there is no module the path we are hitting is not one related to modules, like /
				return next(c)
			}

			// not checking the error. Not all requests include a version
			// i.e. list requests path is like /{module:.+}/@v/list with no version parameter
			version, _ := paths.GetVersion(c)

			if validatorHook, ok := env.ValidatorHook(); ok && version != "" {
				valid, err := validate(validatorHook, mod, version)
				if err != nil {
					entry.SystemErr(err)
					return c.Render(http.StatusInternalServerError, nil)
				}

				if !valid {
					return c.Render(http.StatusForbidden, nil)
				}
			}
			return next(c)
		}
	}
}

func isPseudoVersion(version string) bool {
	return strings.HasPrefix(version, "v0.0.0-")
}

func redirectToOlympusURL(u *url.URL) string {
	return strings.TrimSuffix(env.GetOlympusEndpoint(), "/") + u.Path
}

type validationParams struct {
	Module  string
	Version string
}

func validate(hook, mod, ver string) (bool, error) {
	const op errors.Op = "actions.validate"

	toVal := &validationParams{mod, ver}
	jsonVal, err := json.Marshal(toVal)
	if err != nil {
		return false, errors.E(op, err)
	}

	resp, err := http.Post(hook, "application/json", bytes.NewBuffer(jsonVal))
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
