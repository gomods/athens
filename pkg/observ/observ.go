package observ

import (
	"context"
	"fmt"
	"log"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/gomods/athens/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// RegisterExporter determines the type of TraceExporter service for exporting traces.
// User can choose from multiple tracing services (jaeger, datadog, stackdriver).
// RegisterExporter returns the cleanup function for that particular tracing service.
func RegisterExporter(traceExporter, url, service, env string) (func(), error) {
	const op errors.Op = "observ.RegisterExporter"

	switch traceExporter {
	case "jaeger":
		return registerJaegerExporter(url, service, env)
	case "datadog":
		return registerDatadogExporter(url, service, env)
	case "stackdriver":
		return registerStackdriverExporter(url, env)
	case "":
		return nil, errors.E(op, "Exporter not specified. Traces won't be exported")
	default:
		return nil, errors.E(op, fmt.Sprintf("Exporter %s not supported. Please open PR or an issue at github.com/gomods/athens", traceExporter))
	}
}

// registerJaegerExporter creates an OTLP HTTP exporter for exporting traces to Jaeger.
// Jaeger natively accepts OTLP on port 4318.
func registerJaegerExporter(url, service, env string) (func(), error) {
	const op errors.Op = "observ.registerJaegarExporter"

	if url == "" {
		return nil, errors.E(op, "Exporter URL is empty. Traces won't be exported")
	}

	ctx := context.Background()

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(url),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, errors.E(op, err)
	}

	tp := newTracerProvider(exporter, service, env)

	return shutdownFunc(tp), nil
}

// registerDatadogExporter creates an OTLP HTTP exporter pointed at the DD agent.
// The Datadog Agent accepts OTLP on port 4318.
func registerDatadogExporter(url, service, env string) (func(), error) {
	const op errors.Op = "observ.registerDatadogExporter"

	if url == "" {
		return nil, errors.E(op, "Exporter URL is empty. Traces won't be exported")
	}

	ctx := context.Background()

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(url),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, errors.E(op, err)
	}

	tp := newTracerProvider(exporter, service, env)

	return shutdownFunc(tp), nil
}

func registerStackdriverExporter(projectID, env string) (func(), error) {
	const op errors.Op = "observ.registerStackdriverExporter"

	exporter, err := texporter.New(texporter.WithProjectID(projectID))
	if err != nil {
		return nil, errors.E(op, err)
	}

	tp := newTracerProvider(exporter, "", env)

	return shutdownFunc(tp), nil
}

func newTracerProvider(exporter sdktrace.SpanExporter, service, env string) *sdktrace.TracerProvider {
	r := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(service),
	)

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(r),
	}

	if env == "development" {
		opts = append(opts, sdktrace.WithSampler(sdktrace.AlwaysSample()))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)

	return tp
}

func shutdownFunc(tp *sdktrace.TracerProvider) func() {
	return func() {
		err := tp.Shutdown(context.Background())
		if err != nil {
			log.Printf("failed to shutdown tracer provider: %v", err)
		}
	}
}

// StartSpan takes in a Context Interface and opName and starts a span.
// It returns the new attached context and span.
func StartSpan(ctx context.Context, op string) (context.Context, trace.Span) {
	return otel.Tracer("athens").Start(ctx, op)
}
