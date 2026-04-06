package observ

import (
	"context"
	"testing"
	"time"

	"go.opencensus.io/stats/view"
)

func TestCacheLookupMetric(t *testing.T) {
	// Register only the cache view
	if err := view.Register(cacheLookupView); err != nil {
		t.Fatalf("failed to register view: %v", err)
	}
	defer view.Unregister(cacheLookupView)

	ctx := context.Background()

	RecordCacheLookup(ctx, "hit", "info")

	rows, err := view.RetrieveData("cache_lookup_total")
	if err != nil {
		t.Fatalf("failed to retrieve data: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	count := rows[0].Data.(*view.CountData).Value
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
}

func TestUpstreamFetchCounter(t *testing.T) {
	if err := view.Register(upstreamFetchView); err != nil {
		t.Fatalf("failed to register view: %v", err)
	}
	defer view.Unregister(upstreamFetchView)

	ctx := context.Background()

	RecordUpstreamFetch(ctx, "success")

	rows, err := view.RetrieveData("upstream_fetch_total")
	if err != nil {
		t.Fatalf("failed to retrieve data: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	count := rows[0].Data.(*view.CountData).Value
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
}

func TestUpstreamFetchDurationHistogram(t *testing.T) {
	if err := view.Register(upstreamFetchLatencyView); err != nil {
		t.Fatalf("failed to register view: %v", err)
	}
	defer view.Unregister(upstreamFetchLatencyView)

	ctx := context.Background()

	RecordUpstreamFetchDuration(ctx, "success", 2*time.Second)

	rows, err := view.RetrieveData("upstream_fetch_duration_seconds")
	if err != nil {
		t.Fatalf("failed to retrieve data: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	dist := rows[0].Data.(*view.DistributionData)

	if dist.Count != 1 {
		t.Fatalf("expected count 1, got %d", dist.Count)
	}
}
