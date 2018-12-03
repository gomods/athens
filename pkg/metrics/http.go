package metrics

import (
	"github.com/gobuffalo/buffalo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewHTTPCollectors creates collectors for http metrics
func NewHTTPCollectors() []prometheus.Collector {
	total := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "athens_http_request_total",
			Help: "Total number of request handled by Athens server",
		},
		[]string{"code"},
	)

	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "athens_http_request_duration_seconds",
			Help:    "Duration of request handled by Athens server",
			Buckets: []float64{.1, 1, 5, 10},
		},
		[]string{},
	)

	return []prometheus.Collector{total, duration}
}

// Middleware is middleware that instruments buffalo handlers
func Middleware(collectors ...prometheus.Collector) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(ctx buffalo.Context) error {
			handler := next

			for _, collector := range collectors {
				switch c := collector.(type) {
				case *prometheus.CounterVec:
					handler = buffalo.WrapHandler(promhttp.InstrumentHandlerCounter(c, buffalo.WrapBuffaloHandlerFunc(handler)))
				case *prometheus.HistogramVec:
					handler = buffalo.WrapHandler(promhttp.InstrumentHandlerDuration(c, buffalo.WrapBuffaloHandlerFunc(handler)))
				}
			}

			return handler(ctx)
		}
	}
}
