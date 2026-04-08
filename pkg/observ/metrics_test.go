package observ

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func resetMetrics(t *testing.T) *sdkmetric.ManualReader {
	t.Helper()

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(mp)

	// Reset sync.Once so metrics re-initialize with test provider.
	// Safe: tests run sequentially, each gets its own provider.
	metricsOnce = sync.Once{}

	return reader
}

func collectMetrics(t *testing.T, reader *sdkmetric.ManualReader) metricdata.ResourceMetrics {
	t.Helper()

	var rm metricdata.ResourceMetrics

	err := reader.Collect(context.Background(), &rm)
	if err != nil {
		t.Fatalf("failed to collect metrics: %v", err)
	}

	return rm
}

func findMetric(rm metricdata.ResourceMetrics, name string) *metricdata.Metrics {
	for _, sm := range rm.ScopeMetrics {
		for i := range sm.Metrics {
			if sm.Metrics[i].Name == name {
				return &sm.Metrics[i]
			}
		}
	}

	return nil
}

func TestCacheLookupMetric(t *testing.T) {
	reader := resetMetrics(t)
	ctx := context.Background()

	RecordCacheLookup(ctx, "hit", "info")

	rm := collectMetrics(t, reader)
	m := findMetric(rm, "cache_lookup_total")

	if m == nil {
		t.Fatal("expected cache_lookup_total metric")
	}

	sum, ok := m.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatal("expected Sum[int64] data type")
	}

	if len(sum.DataPoints) != 1 {
		t.Fatalf("expected 1 data point, got %d", len(sum.DataPoints))
	}

	if sum.DataPoints[0].Value != 1 {
		t.Fatalf("expected count 1, got %d", sum.DataPoints[0].Value)
	}
}

func TestUpstreamFetchCounter(t *testing.T) {
	reader := resetMetrics(t)
	ctx := context.Background()

	RecordUpstreamFetch(ctx, "success")

	rm := collectMetrics(t, reader)
	m := findMetric(rm, "upstream_fetch_total")

	if m == nil {
		t.Fatal("expected upstream_fetch_total metric")
	}

	sum, ok := m.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatal("expected Sum[int64] data type")
	}

	if len(sum.DataPoints) != 1 {
		t.Fatalf("expected 1 data point, got %d", len(sum.DataPoints))
	}

	if sum.DataPoints[0].Value != 1 {
		t.Fatalf("expected count 1, got %d", sum.DataPoints[0].Value)
	}
}

func TestUpstreamFetchDurationHistogram(t *testing.T) {
	reader := resetMetrics(t)
	ctx := context.Background()

	RecordUpstreamFetchDuration(ctx, "success", 2*time.Second)

	rm := collectMetrics(t, reader)
	m := findMetric(rm, "upstream_fetch_duration_seconds")

	if m == nil {
		t.Fatal("expected upstream_fetch_duration_seconds metric")
	}

	hist, ok := m.Data.(metricdata.Histogram[float64])
	if !ok {
		t.Fatal("expected Histogram[float64] data type")
	}

	if len(hist.DataPoints) != 1 {
		t.Fatalf("expected 1 data point, got %d", len(hist.DataPoints))
	}

	if hist.DataPoints[0].Count != 1 {
		t.Fatalf("expected count 1, got %d", hist.DataPoints[0].Count)
	}
}
