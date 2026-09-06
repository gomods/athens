package observ

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/otlptranslator"
	"go.opentelemetry.io/otel"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

// setupTestMetrics installs a MeterProvider backed by a fresh Prometheus registry
// and initializes Athens' custom instruments, mirroring registerPrometheusExporter.
// It returns the registry so tests can scrape the recorded values.
func setupTestMetrics(t *testing.T) *prometheus.Registry {
	t.Helper()

	registry := prometheus.NewRegistry()
	exporter, err := otelprom.New(
		otelprom.WithRegisterer(registry),
		otelprom.WithNamespace("proxy"),
		otelprom.WithTranslationStrategy(otlptranslator.UnderscoreEscapingWithoutSuffixes),
		otelprom.WithoutScopeInfo(),
	)
	if err != nil {
		t.Fatalf("failed to create prometheus exporter: %v", err)
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	otel.SetMeterProvider(provider)
	t.Cleanup(func() { _ = provider.Shutdown(t.Context()) })

	if err := initMetrics(); err != nil {
		t.Fatalf("failed to init metrics: %v", err)
	}

	return registry
}

// findMetricFamily returns the gathered metric family with the given name, or nil.
func findMetricFamily(t *testing.T, registry *prometheus.Registry, name string) *dto.MetricFamily {
	t.Helper()

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
	for _, fam := range families {
		if fam.GetName() == name {
			return fam
		}
	}
	return nil
}

func TestCacheLookupMetric(t *testing.T) {
	registry := setupTestMetrics(t)

	RecordCacheLookup(t.Context(), "hit", "info")

	fam := findMetricFamily(t, registry, "proxy_cache_lookup_total")
	if fam == nil {
		t.Fatal("expected metric family proxy_cache_lookup_total to be present")
	}
	if got := len(fam.GetMetric()); got != 1 {
		t.Fatalf("expected 1 metric, got %d", got)
	}
	if got := fam.GetMetric()[0].GetCounter().GetValue(); got != 1 {
		t.Fatalf("expected counter value 1, got %v", got)
	}
}

func TestUpstreamFetchCounter(t *testing.T) {
	registry := setupTestMetrics(t)

	RecordUpstreamFetch(t.Context(), "success")

	fam := findMetricFamily(t, registry, "proxy_upstream_fetch_total")
	if fam == nil {
		t.Fatal("expected metric family proxy_upstream_fetch_total to be present")
	}
	if got := fam.GetMetric()[0].GetCounter().GetValue(); got != 1 {
		t.Fatalf("expected counter value 1, got %v", got)
	}
}

func TestUpstreamFetchDurationHistogram(t *testing.T) {
	registry := setupTestMetrics(t)

	RecordUpstreamFetchDuration(t.Context(), "success", 2*time.Second)

	fam := findMetricFamily(t, registry, "proxy_upstream_fetch_duration_seconds")
	if fam == nil {
		t.Fatal("expected metric family proxy_upstream_fetch_duration_seconds to be present")
	}
	hist := fam.GetMetric()[0].GetHistogram()
	if got := hist.GetSampleCount(); got != 1 {
		t.Fatalf("expected sample count 1, got %d", got)
	}
	if got := hist.GetSampleSum(); got != 2 {
		t.Fatalf("expected sample sum 2, got %v", got)
	}
}
