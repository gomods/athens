package observ

import (
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/errors"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
)

// RegisterStatsExporter determines the type of StatsExporter service for exporting stats from Opencensus
// Currently it supports: prometheus
func RegisterStatsExporter(app *buffalo.App, statsExporter, service string) (func(), error) {
	const op errors.Op = "observ.RegisterStatsExporter"

	switch statsExporter {
	case "prometheus":
		if err := registerPrometheusExporter(app, service); err != nil {
			return nil, errors.E(op, err)
		}
	case "":
		return nil, errors.E(op, "StatsExporter not specified. Stats won't be collected")
	default:
		return nil, errors.E(op, fmt.Sprintf("StatsExporter %s not supported. Please open PR or an issue at github.com/gomods/athens", statsExporter))
	}

	if err := registerViews(); err != nil {
		return nil, errors.E(op, err)
	}

	// Currently func() prop it's not needed by Prometheus exporter.
	// This param should be used to pass StackDriver Flush or
	// DataDog Stop method when this Exporters will be implemented.
	return func() {}, nil
}

// StatsMiddleware is middleware that instruments buffalo handlers
func StatsMiddleware() buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(ctx buffalo.Context) error {
			och := &ochttp.Handler{
				Handler: buffalo.WrapBuffaloHandlerFunc(next),
			}

			return buffalo.WrapHandler(och)(ctx)
		}
	}
}

// registerViews register stats which should be collected in Athens
func registerViews() error {
	const op errors.Op = "observ.registerViews"
	if err := view.Register(ochttp.DefaultServerViews...); err != nil {
		return errors.E(op, err)
	}

	return nil
}

// registerPrometheusExporter creates exporter that collects stats for Prometheus.
func registerPrometheusExporter(app *buffalo.App, service string) error {
	const op errors.Op = "observ.registerPrometheusExporter"

	prom, err := prometheus.NewExporter(prometheus.Options{
		Namespace: service,
	})
	if err != nil {
		return errors.E(op, err)
	}

	app.GET("/metrics", metricsHandler(prom))
	app.Middleware.Skip(StatsMiddleware(), metricsHandler(prom))

	view.RegisterExporter(prom)

	return nil
}

func metricsHandler(handler http.Handler) buffalo.Handler {
	return buffalo.WrapHandler(handler)
}
