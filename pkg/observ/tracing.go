package observ

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/config/env"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
)

// ENV is used to define the sampling rate
var ENV = env.GoEnvironmentWithDefault("development")

// observabilityContext is a private context that is used by the packages to start the span
type observabilityContext struct {
	buffalo.Context
	spanCtx context.Context
}

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

	fmt.Println(ENV, "---------------------------")
	if ENV == "development" {
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	}

	return je
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

			span.AddAttributes(
				requestAttrs(ctx.Request())...,
			)

			return next(&observabilityContext{Context: ctx, spanCtx: spanCtx})
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
