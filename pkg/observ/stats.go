package observ

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// RegisterStatsExporter determines the type of StatsExporter service for exporting stats.
// Currently it supports: prometheus, stackdriver, datadog.
func RegisterStatsExporter(r *mux.Router, statsExporter, service string) (func(), error) {
	const op errors.Op = "observ.RegisterStatsExporter"

	switch statsExporter {
	case "prometheus":
		cleanup, err := registerPrometheusExporter(r, service)
		if err != nil {
			return nil, errors.E(op, err)
		}

		return cleanup, nil
	case "stackdriver":
		cleanup, err := registerStatsStackDriverExporter(service)
		if err != nil {
			return nil, errors.E(op, err)
		}

		return cleanup, nil
	case "datadog":
		cleanup, err := registerStatsDataDogExporter(service)
		if err != nil {
			return nil, errors.E(op, err)
		}

		return cleanup, nil
	case "":
		return nil, errors.E(op, "StatsExporter not specified. Stats won't be collected")
	default:
		return nil, errors.E(op, fmt.Sprintf("StatsExporter %s not supported. Please open PR or an issue at github.com/gomods/athens", statsExporter))
	}
}

// registerPrometheusExporter creates exporter that collects stats for Prometheus.
func registerPrometheusExporter(r *mux.Router, service string) (func(), error) {
	const op errors.Op = "observ.registerPrometheusExporter"

	reader, err := promexporter.New()
	if err != nil {
		return nil, errors.E(op, err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(service),
	)

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader), sdkmetric.WithResource(res))
	otel.SetMeterProvider(mp)

	r.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)

	return shutdownMeterProvider(mp), nil
}

func registerStatsDataDogExporter(service string) (func(), error) {
	const op errors.Op = "observ.registerStatsDataDogExporter"

	exporter, err := otlpmetrichttp.New(context.Background(),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, errors.E(op, err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(service),
	)

	reader := sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(60*time.Second))
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader), sdkmetric.WithResource(res))
	otel.SetMeterProvider(mp)

	return shutdownMeterProvider(mp), nil
}

func registerStatsStackDriverExporter(projectID string) (func(), error) {
	const op errors.Op = "observ.registerStatsStackDriverExporter"

	exporter, err := mexporter.New(mexporter.WithProjectID(projectID))
	if err != nil {
		return nil, errors.E(op, err)
	}

	reader := sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(60*time.Second))
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(mp)

	return shutdownMeterProvider(mp), nil
}

func shutdownMeterProvider(mp *sdkmetric.MeterProvider) func() {
	return func() {
		err := mp.Shutdown(context.Background())
		if err != nil {
			log.Printf("failed to shutdown meter provider: %v", err)
		}
	}
}
