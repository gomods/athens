package observ

import (
	"context"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	cacheResult = tag.MustNewKey("cache_result")
	cacheType   = tag.MustNewKey("cache_type")
	fetchResult = tag.MustNewKey("fetch_result")
)

var (
	cacheStats                 = stats.Int64("cache_lookup_total", "Count of cache lookup results", stats.UnitDimensionless)
	upstreamFetchStats         = stats.Int64("upstream_fetch_total", "Count of upstream fetch attempts", stats.UnitDimensionless)
	upstreamFetchDurationStats = stats.Float64("upstream_fetch_duration_seconds", "Distribution of upstream fetch latency in seconds", stats.UnitSeconds)
)

var upstreamExponentialBuckets = []float64{0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30}

var (
	cacheLookupView = &view.View{
		Name:        "cache_lookup_total",
		Measure:     cacheStats,
		Description: "Count of cache lookup results",
		TagKeys:     []tag.Key{cacheResult, cacheType},
		Aggregation: view.Count(),
	}
	upstreamFetchView = &view.View{
		Name:        "upstream_fetch_total",
		Measure:     upstreamFetchStats,
		Description: "Count of upstream fetch attempts",
		TagKeys:     []tag.Key{fetchResult},
		Aggregation: view.Count(),
	}
	upstreamFetchLatencyView = &view.View{
		Name:        "upstream_fetch_duration_seconds",
		Measure:     upstreamFetchDurationStats,
		Description: "Distribution of upstream fetch latency in seconds",
		TagKeys:     []tag.Key{fetchResult},
		Aggregation: view.Distribution(upstreamExponentialBuckets...),
	}
)

func customViews() []*view.View {
	return []*view.View{cacheLookupView, upstreamFetchView, upstreamFetchLatencyView}
}

func RecordCacheLookup(ctx context.Context, result, typ string) {
	ctx, _ = tag.New(ctx,
		tag.Insert(cacheResult, result),
		tag.Insert(cacheType, typ),
	)
	stats.Record(ctx, cacheStats.M(1))
}

func RecordUpstreamFetch(ctx context.Context, result string) {
	ctx, _ = tag.New(ctx,
		tag.Insert(fetchResult, result),
	)
	stats.Record(ctx, upstreamFetchStats.M(1))
}

func RecordUpstreamFetchDuration(ctx context.Context, result string, duration time.Duration) {
	ctx, _ = tag.New(ctx,
		tag.Insert(fetchResult, result),
	)
	stats.Record(ctx, upstreamFetchDurationStats.M(duration.Seconds()))
}
