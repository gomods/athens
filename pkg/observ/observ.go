package observ

import (
	"context"
	"fmt"

	"contrib.go.opencensus.io/exporter/stackdriver"
	datadog "github.com/DataDog/opencensus-go-exporter-datadog"
	"github.com/gomods/athens/pkg/errors"
	"contrib.go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
)

// RegisterExporter determines the type of TraceExporter service for exporting traces from opencensus
// User can choose from multiple tracing services (datadog, jaegar)
// RegisterExporter returns the 'Flush' function for that particular tracing service
func RegisterExporter(traceExporter, URL, service, ENV string) (func(), error) {
	const op errors.Op = "observ.RegisterExporter"
	switch traceExporter {
	case "jaeger":
		return registerJaegerExporter(URL, service, ENV)
	case "datadog":
		return registerDatadogExporter(URL, service, ENV)
	case "stackdriver":
		return registerStackdriverExporter(URL, ENV)
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
	const op errors.Op = "observ.registerJaegarExporter"
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
	const op errors.Op = "observ.registerStackdriverExporter"
	ex, err := stackdriver.NewExporter(stackdriver.Options{ProjectID: projectID})
	if err != nil {
		return nil, errors.E(op, err)
	}
	traceRegisterExporter(ex, ENV)
	return ex.Flush, nil
}

// StartSpan takes in a Context Interface and opName and starts a span. It returns the new attached ObserverContext
// and span
func StartSpan(ctx context.Context, op string) (context.Context, *trace.Span) {
	return trace.StartSpan(ctx, op)
}
