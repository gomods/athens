package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func addMetrics(app *buffalo.App, skip ...buffalo.Handler) {
	http := metrics.NewHTTPCollectors()

	registry := prometheus.NewRegistry()
	registry.MustRegister(http...)

	app.Use(metrics.Middleware(http...))
	app.GET("/metrics", metricsHandler(registry))

	app.Middleware.Skip(metrics.Middleware(), skip...)
	app.Middleware.Skip(metrics.Middleware(), metricsHandler(registry))
}

func metricsHandler(registry *prometheus.Registry) buffalo.Handler {
	return buffalo.WrapHandler(promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{},
	))
}
