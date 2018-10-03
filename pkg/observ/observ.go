package observ

import (
	"context"
	"net/http"

	datadog "github.com/DataDog/opencensus-go-exporter-datadog"
	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/errors"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

// observabilityContext is a private context that is used by the packages to start the span
type observabilityContext struct {
	buffalo.Context
	spanCtx context.Context
}

// jaegerTraceExporter is used to hold the trace exporter for Jaeger
type jaegerTraceExporter struct {
	exporter (*(jaeger.Exporter))
}

// datadogTraceExporter is used to hold the trace exporter for Datadog
type datadogTraceExporter struct {
	exporter (*(datadog.Exporter))
}

// Exporter might be a JaejerExporter, DataDogExporter, etc
// All Exporters *must* implement these two functions.
// Flush method stops the trace exporter, this method needs to be
// called for most of the TraceExporters using opencensus
type Exporter interface {
	RegisterTraceExporter(URL, service, ENV string) error
	Flush()
}

// RetrieveTracer determines the type of TraceExporter service for exporting traces from opencensus
// User can choose from multiple tracing services (datadog, jaegar)
// TODO: Extend beyond jaeger and datadog
func RetrieveTracer(TraceExporter string) (Exporter, error) {

	const op errors.Op = "RegisterTracer"

	switch TraceExporter {
	case "jaeger":
		return jaegerTraceExporter{}, nil
	case "datadog":
		return datadogTraceExporter{}, nil
	default:
		return nil, errors.E(op, "Exporter not specified. Traces won't be exported")
	}
}

func (j jaegerTraceExporter) Flush() {
	j.exporter.Flush()
}

func (d datadogTraceExporter) Flush() {
	d.exporter.Stop()
}

// RegisterTraceExporter creates a jaeger exporter for exporting traces to opencensus.
// Currently uses the 'TraceExporter' variable in the config file.
// It should in the future have a nice sampling rate defined
func (j jaegerTraceExporter) RegisterTraceExporter(URL, service, ENV string) error {
	const op errors.Op = "RegisterTracer"

	if URL == "" {
		return errors.E(op, "Exporter URL is empty. Traces won't be exported")
	}

	var err error
	j.exporter, err = jaeger.NewExporter(jaeger.Options{
		Endpoint: URL,
		Process: jaeger.Process{
			ServiceName: service,
			Tags: []jaeger.Tag{
				// IP Tag ensures Jaeger's clock isn't skewed.
				// If/when we have traces across different servers,
				// we should make this IP dynamic.
				jaeger.StringTag("ip", "127.0.0.1"),
			},
		},
	})

	if err != nil {
		return errors.E(op, err)
	}

	// And now finally register it as a Trace Exporter
	trace.RegisterExporter(j.exporter)
	if ENV == "development" {
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	}

	return nil
}

// RegisterDatadogTracerExporter creates a datadog exporter.
func (d datadogTraceExporter) RegisterTraceExporter(URL, service, ENV string) error {

	d.exporter = datadog.NewExporter(
		datadog.Options{
			TraceAddr: URL,
			Service:   service,
		})

	trace.RegisterExporter(d.exporter)

	// For demoing purposes, always sample. In a production application, you should
	// configure this to a trace.ProbabilitySampler set at the desired
	// probability.
	if ENV == "development" {
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	}
	return nil

}

// Tracer is a middleware that starts a span from the top of a buffalo context
// and propates it to the bottom of the stack
func Tracer(service string) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(ctx buffalo.Context) error {
			spanCtx, span := trace.StartSpan(ctx,
				ctx.Request().URL.Path,
				trace.WithSpanKind(trace.SpanKindClient))
			defer span.End()

			handler := next(&observabilityContext{Context: ctx, spanCtx: spanCtx})

			// Add request attributes
			span.AddAttributes(
				requestAttrs(ctx.Request())...,
			)

			// SetSpan Status from response
			if resp, ok := ctx.Response().(*buffalo.Response); ok {
				span.SetStatus(ochttp.TraceStatus(resp.Status, ""))
				span.AddAttributes(trace.Int64Attribute("http.status_code", int64(resp.Status)))
			}

			return handler
		}
	}
}

// Applies request information to the span
func requestAttrs(r *http.Request) []trace.Attribute {
	// From: https://github.com/census-instrumentation/opencensus-go/blob/master/plugin/ochttp/trace.go
	return []trace.Attribute{
		trace.StringAttribute("http.path", r.URL.Path),
		trace.StringAttribute("http.host", r.URL.Host),
		trace.StringAttribute("http.method", r.Method),
		trace.StringAttribute("http.user_agent", r.UserAgent()),
	}
}

// StartSpan takes in a Context Interface and opName and starts a span. It returns the new attached ObserverContext
// and span
func StartSpan(ctx context.Context, op string) (context.Context, *trace.Span) {
	oCtx, ok := ctx.(*observabilityContext)
	if ok {
		return trace.StartSpan(oCtx.spanCtx, op)
	}
	return trace.StartSpan(ctx, op)
}
