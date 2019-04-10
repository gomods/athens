package observ

import (
	"fmt"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	datadog "github.com/DataDog/opencensus-go-exporter-datadog"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gorilla/mux"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
)

// RegisterStatsExporter determines the type of StatsExporter service for exporting stats from Opencensus
// Currently it supports: prometheus
func RegisterStatsExporter(r *mux.Router, statsExporter, service string) (func(), error) {
	const op errors.Op = "observ.RegisterStatsExporter"
	var stop = func() {}
	var err error
	switch statsExporter {
	case "prometheus":
		if err := registerPrometheusExporter(r, service); err != nil {
			return nil, errors.E(op, err)
		}
	case "stackdriver":
		if stop, err = registerStatsStackDriverExporter(service); err != nil {
			return nil, errors.E(op, err)
		}
	case "datadog":
		if stop, err = registerStatsDataDogExporter(service); err != nil {
			return nil, errors.E(op, err)
		}
	case "":
		return nil, errors.E(op, "StatsExporter not specified. Stats won't be collected")
	default:
		return nil, errors.E(op, fmt.Sprintf("StatsExporter %s not supported. Please open PR or an issue at github.com/gomods/athens", statsExporter))
	}
	if err := registerViews(); err != nil {
		return nil, errors.E(op, err)
	}

	return stop, nil
}

// registerPrometheusExporter creates exporter that collects stats for Prometheus.
func registerPrometheusExporter(r *mux.Router, service string) error {
	const op errors.Op = "observ.registerPrometheusExporter"
	prom, err := prometheus.NewExporter(prometheus.Options{
		Namespace: service,
	})
	if err != nil {
		return errors.E(op, err)
	}

	r.Handle("/metrics", prom).Methods(http.MethodGet)

	view.RegisterExporter(prom)

	return nil
}

func registerStatsDataDogExporter(service string) (func(), error) {
	const op errors.Op = "observ.registerStatsDataDogExporter"

	dd := datadog.NewExporter(datadog.Options{Service: service})
	if dd == nil {
		return nil, errors.E(op, "Failed to initialize data dog exporter")
	}
	view.RegisterExporter(dd)
	return dd.Stop, nil
}

func registerStatsStackDriverExporter(projectID string) (func(), error) {
	const op errors.Op = "observ.registerStatsStackDriverExporter"

	sd, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: projectID,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}

	view.RegisterExporter(sd)
	view.SetReportingPeriod(60 * time.Second)

	return sd.Flush, nil
}

// registerViews register stats which should be collected in Athens
func registerViews() error {
	const op errors.Op = "observ.registerViews"
	if err := view.Register(ochttp.DefaultServerViews...); err != nil {
		return errors.E(op, err)
	}

	return nil
}
