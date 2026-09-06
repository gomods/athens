package observ

import (
	"context"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Attribute keys used on Athens' custom metrics.
const (
	attrCacheResult = "cache_result"
	attrCacheType   = "cache_type"
	attrFetchResult = "fetch_result"
)

// upstreamExponentialBuckets are the histogram boundaries (in seconds) for
// upstream fetch latency.
var upstreamExponentialBuckets = []float64{0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30}

// Custom instruments. They are nil until initMetrics runs (i.e. when the stats
// exporter is registered); the Record* helpers guard against that so recording
// is a no-op when metrics are disabled, mirroring OpenCensus' behavior.
var (
	cacheLookupCounter    metric.Int64Counter
	upstreamFetchCounter  metric.Int64Counter
	upstreamFetchDuration metric.Float64Histogram
)

// initMetrics creates Athens' custom instruments from the global MeterProvider.
// It must be called after the MeterProvider has been installed.
func initMetrics() error {
	const op errors.Op = "observ.initMetrics"
	meter := otel.Meter(instrumentationName)

	var err error
	cacheLookupCounter, err = meter.Int64Counter(
		"cache_lookup_total",
		metric.WithDescription("Count of cache lookup results"),
	)
	if err != nil {
		return errors.E(op, err)
	}

	upstreamFetchCounter, err = meter.Int64Counter(
		"upstream_fetch_total",
		metric.WithDescription("Count of upstream fetch attempts"),
	)
	if err != nil {
		return errors.E(op, err)
	}

	upstreamFetchDuration, err = meter.Float64Histogram(
		"upstream_fetch_duration_seconds",
		metric.WithDescription("Distribution of upstream fetch latency in seconds"),
		metric.WithExplicitBucketBoundaries(upstreamExponentialBuckets...),
	)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func RecordCacheLookup(ctx context.Context, result, typ string) {
	if cacheLookupCounter == nil {
		return
	}
	cacheLookupCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String(attrCacheResult, result),
		attribute.String(attrCacheType, typ),
	))
}

func RecordUpstreamFetch(ctx context.Context, result string) {
	if upstreamFetchCounter == nil {
		return
	}
	upstreamFetchCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String(attrFetchResult, result),
	))
}

func RecordUpstreamFetchDuration(ctx context.Context, result string, duration time.Duration) {
	if upstreamFetchDuration == nil {
		return
	}
	upstreamFetchDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String(attrFetchResult, result),
	))
}
