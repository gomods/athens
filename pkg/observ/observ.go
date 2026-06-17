package observ

import (
	"context"
	"fmt"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// instrumentationName is the name of the tracer and meter instrumentation scope.
const instrumentationName = "github.com/gomods/athens"

// shutdownTimeout bounds how long provider shutdown may block while flushing.
const shutdownTimeout = 5 * time.Second

// RegisterExporter configures the OpenTelemetry TracerProvider used to export traces.
//
// Athens exports traces via OTLP. The traceExporter argument acts as a toggle:
// "otlp" enables export, an empty value disables it. The url argument, when set,
// overrides the OTLP endpoint; otherwise the standard OTEL_EXPORTER_OTLP_* environment
// variables are honored. samplingFraction is the fraction of root traces sampled outside
// of development (0 samples nothing, 1 samples everything). RegisterExporter returns a
// cleanup function that flushes and shuts down the provider; the caller is responsible for
// calling it at shutdown.
func RegisterExporter(traceExporter, url, service, env string, samplingFraction float64) (func(), error) {
	const op errors.Op = "observ.RegisterExporter"
	switch traceExporter {
	case "otlp":
		return registerOTLPExporter(url, service, env, samplingFraction)
	case "jaeger", "datadog", "stackdriver":
		return nil, errors.E(op, fmt.Sprintf(
			"Exporter %q is no longer supported. Athens now exports traces via OpenTelemetry (OTLP). "+
				"Set ATHENS_TRACE_EXPORTER=otlp and point ATHENS_TRACE_EXPORTER_URL "+
				"(or OTEL_EXPORTER_OTLP_ENDPOINT) at an OTLP collector.", traceExporter))
	case "":
		return nil, errors.E(op, "Exporter not specified. Traces won't be exported")
	default:
		return nil, errors.E(op, fmt.Sprintf("Exporter %s not supported. Please open PR or an issue at github.com/gomods/athens", traceExporter))
	}
}

// registerOTLPExporter creates an OTLP gRPC trace exporter and installs a global
// TracerProvider. In development everything is sampled; otherwise root spans are sampled
// at samplingFraction and child spans follow their parent's decision.
func registerOTLPExporter(url, service, env string, samplingFraction float64) (func(), error) {
	const op errors.Op = "observ.registerOTLPExporter"
	ctx := context.Background()

	opts := []otlptracegrpc.Option{otlptracegrpc.WithInsecure()}
	if url != "" {
		opts = append(opts, otlptracegrpc.WithEndpointURL(url))
	}
	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, errors.E(op, err)
	}

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(service)))
	if err != nil {
		return nil, errors.E(op, err)
	}

	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(samplingFraction))
	if env == "development" {
		sampler = sdktrace.AlwaysSample()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = tp.Shutdown(ctx)
	}, nil
}

// StartSpan takes in a Context and opName and starts a span. It returns the new
// context carrying the span and the span itself.
func StartSpan(ctx context.Context, op string) (context.Context, oteltrace.Span) {
	return otel.Tracer(instrumentationName).Start(ctx, op)
}
