package observ

import (
	"context"
	"fmt"
	"net/http"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/ocagent"
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

// RegisterExporter determines the type of TraceExporter service for exporting traces from opencensus
// User can choose from multiple tracing services (datadog, jaegar)
// RegisterExporter returns the 'Flush' function for that particular tracing service
func RegisterExporter(traceExporter, URL, service, ENV string) (func(), error) {
	const op errors.Op = "RegisterExporter"
	switch traceExporter {
	case "jaeger":
		return registerJaegerExporter(URL, service, ENV)
	case "datadog":
		return registerDatadogExporter(URL, service, ENV)
	case "stackdriver":
		return registerStackdriverExporter(URL, ENV)
	case "appinsights":
		return registerAppInsightsExporter(URL, ENV)
	case "":
		return nil, errors.E(op, "Exporter not specified. Traces won't be exported")
	default:
		return nil, errors.E(op, fmt.Sprintf("Exporter %s not supported. Please open PR or an issue at github.com/gomods/athens", traceExporter))
	}
}

// registerJaegerExporter creates a jaeger exporter for exporting traces to opencensus.
// Currently uses the 'TraceExporter' variable in the config file.
// It should in the future have a nice sampling rate defined
func registerJaegerExporter(URL, service, ENV string) (func(), error) {
	const op errors.Op = "registerJaegarExporter"
	if URL == "" {
		return nil, errors.E(op, "Exporter URL is empty. Traces won't be exported")
	}
	ex, err := jaeger.NewExporter(jaeger.Options{
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
		return nil, errors.E(op, err)
	}
	traceRegisterExporter(ex, ENV)
	return ex.Flush, nil
}

func traceRegisterExporter(exporter trace.Exporter, ENV string) {
	trace.RegisterExporter(exporter)
	if ENV == "development" {
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	}
}

// registerDatadogTracerExporter creates a datadog exporter.
// Currently uses the 'TraceExporter' variable in the config file.
func registerDatadogExporter(URL, service, ENV string) (func(), error) {
	ex := datadog.NewExporter(
		datadog.Options{
			TraceAddr: URL,
			Service:   service,
		})
	traceRegisterExporter(ex, ENV)
	return ex.Stop, nil
}

func registerStackdriverExporter(projectID, ENV string) (func(), error) {
	const op errors.Op = "registerStackdriverExporter"
	ex, err := stackdriver.NewExporter(stackdriver.Options{ProjectID: projectID})
	if err != nil {
		return nil, errors.E(op, err)
	}
	traceRegisterExporter(ex, ENV)
	return ex.Flush, nil
}

func registerAppInsightsExporter(svcName, endpoint, ENV string) (func(), error) {
	const op errors.Op = "registerAppInsightsExporter"
	ex, err := ocagent.NewExporter(
		ocagent.WithInsecure(),
		ocagent.WithServiceName(svcName),
		ocagent.WithAddress(endpoint),
	)
	traceRegisterExporter(ex, ENV)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return ex.Flush, nil
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
