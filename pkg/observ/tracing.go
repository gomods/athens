package observ

import (
	"context"
	"log"

	"github.com/gobuffalo/buffalo"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
)

// TraceCtx is the name of the key where we will store the traceCtx in buffalo context
const TraceCtx = "traceCtx"

// RegisterTraceExporter returns a jaeger exporter for exporting traces to opencensus.
// It should in the future have a nice sampling rate defined
func RegisterTraceExporter(service string) *(jaeger.Exporter) {
	collectorEndpointURI := "http://0.0.0.0:14268"

	je, err := jaeger.NewExporter(jaeger.Options{
		Endpoint:    collectorEndpointURI,
		ServiceName: service,
	})
	if err != nil {
		log.Fatalf("Failed to create the Jaeger exporter: %v", err)
	}

	// And now finally register it as a Trace Exporter
	trace.RegisterExporter(je)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	return je
}

// Tracer is a middleware that starts a span from the top of a buffalo context
// and propates it to the bottom of the stack
func Tracer(service string) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(ctx buffalo.Context) error {
			var spanCtx context.Context
			var span *trace.Span
			if tCtx := ctx.Data()[TraceCtx]; tCtx != nil {
				spanCtx, span = trace.StartSpan(tCtx.(context.Context), service)
				defer span.End()
			} else {
				spanCtx, span = trace.StartSpan(ctx, service)
				defer span.End()

				// Set the traceCtx in spanCtx
				ctx.Set("traceCtx", spanCtx)
			}

			// Add attributes that are required
			span.AddAttributes(trace.StringAttribute("url", ctx.Request().URL.String()))

			return next(ctx)
		}
	}
}

// StartSpan takes in a Context Interface and opName and starts a span. It returns the new attached ObserverContext
// and span
func StartSpan(ctx context.Context, op string) (context.Context, *trace.Span) {
	return trace.StartSpan(ctx, op)
}

// StartBuffaloSpan takes in a BuffaloContext Interface and opName and starts a span. It returns the new attached ObserverContext
// and span
func StartBuffaloSpan(ctx buffalo.Context, op string) (buffalo.Context, context.Context, *trace.Span) {
	var spanCtx context.Context
	var span *trace.Span
	// StartSpan if the span already exists
	if tCtx := ctx.Data()[TraceCtx]; tCtx != nil {
		spanCtx, span = trace.StartSpan(tCtx.(context.Context), op)
	} else {
		spanCtx, span = trace.StartSpan(ctx, op)
		// Set the traceCtx in spanCtx
		ctx.Set("traceCtx", spanCtx)
	}

	return ctx, spanCtx, span
}
