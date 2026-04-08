package observ

import (
	"context"
	"log"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	cacheLookupCounter    metric.Int64Counter
	upstreamFetchCounter  metric.Int64Counter
	upstreamFetchDuration metric.Float64Histogram
	metricsOnce           sync.Once
)

func initMetrics() {
	metricsOnce.Do(func() {
		meter := otel.Meter("athens")

		var err error

		cacheLookupCounter, err = meter.Int64Counter("cache_lookup_total",
			metric.WithDescription("Count of cache lookup results"),
		)
		if err != nil {
			log.Fatalf("failed to create metric: %v", err)
		}

		upstreamFetchCounter, err = meter.Int64Counter("upstream_fetch_total",
			metric.WithDescription("Count of upstream fetch attempts"),
		)
		if err != nil {
			log.Fatalf("failed to create metric: %v", err)
		}

		upstreamFetchDuration, err = meter.Float64Histogram("upstream_fetch_duration_seconds",
			metric.WithDescription("Distribution of upstream fetch latency in seconds"),
			metric.WithExplicitBucketBoundaries(0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30),
		)
		if err != nil {
			log.Fatalf("failed to create metric: %v", err)
		}
	})
}

// RecordCacheLookup records a cache lookup event with the given result and type.
func RecordCacheLookup(ctx context.Context, result, typ string) {
	initMetrics()
	cacheLookupCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("cache_result", result),
			attribute.String("cache_type", typ),
		),
	)
}

// RecordUpstreamFetch records an upstream fetch event with the given result.
func RecordUpstreamFetch(ctx context.Context, result string) {
	initMetrics()
	upstreamFetchCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("fetch_result", result),
		),
	)
}

// RecordUpstreamFetchDuration records the duration of an upstream fetch.
func RecordUpstreamFetchDuration(ctx context.Context, result string, duration time.Duration) {
	initMetrics()
	upstreamFetchDuration.Record(ctx, duration.Seconds(),
		metric.WithAttributes(
			attribute.String("fetch_result", result),
		),
	)
}
