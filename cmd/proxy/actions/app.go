package actions

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/log"
	mw "github.com/gomods/athens/pkg/middleware"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gorilla/mux"
	"github.com/unrolled/secure"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Service is the name of the service that we want to tag our processes with.
const Service = "proxy"

// App is where all routes and middleware for the proxy
// should be defined. This is the nerve center of your
// application.
//
// App returns the HTTP handler, a cleanup function that should be called
// when the server is shutting down (to flush and stop exporters), and an error.
func App(logger *log.Logger, conf *config.Config) (http.Handler, func(), error) {
	noop := func() {}

	if conf.GithubToken != "" {
		if conf.NETRCPath != "" {
			return nil, noop, fmt.Errorf("cannot provide both GithubToken and NETRCPath")
		}

		err := netrcFromToken(conf.GithubToken)
		if err != nil {
			return nil, noop, fmt.Errorf("creating netrc from token: %w", err)
		}
	}

	// mount .netrc to home dir
	// to have access to private repos.
	err := initializeAuthFile(conf.NETRCPath)
	if err != nil {
		return nil, noop, fmt.Errorf("initializing auth file from netrc: %w", err)
	}

	// mount .hgrc to home dir
	// to have access to private repos.
	err = initializeAuthFile(conf.HGRCPath)
	if err != nil {
		return nil, noop, fmt.Errorf("initializing auth file from hgrc: %w", err)
	}

	r := mux.NewRouter()
	r.Use(
		mw.WithRequestID,
		mw.LogEntryMiddleware(logger),
		mw.RequestLogger,
		secure.New(secure.Options{
			SSLRedirect:     conf.ForceSSL,
			SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
		}).Handler,
	)

	var subRouter *mux.Router

	if prefix := conf.PathPrefix; prefix != "" {
		// certain Ingress Controllers (such as GCP Load Balancer)
		// can not send custom headers and therefore if the proxy
		// is running behind a prefix as well as some authentication
		// mechanism, we should allow the plain / to return 200.
		r.HandleFunc("/", healthHandler).Methods(http.MethodGet)
		subRouter = r.PathPrefix(prefix).Subrouter()
	}

	// RegisterExporter will register an exporter where we will export our traces to.
	// The error from the RegisterExporter would be nil if the tracer was specified by
	// the user and the trace exporter was created successfully.
	// RegisterExporter returns the cleanup function that flushes remaining traces
	// and stops the exporter. The caller is responsible for calling it at shutdown.
	cleanupTraces := noop

	flushTraces, err := observ.RegisterExporter(
		conf.TraceExporter,
		conf.TraceExporterURL,
		Service,
		conf.GoEnv,
	)
	if err != nil {
		logger.Info(err)
	} else {
		cleanupTraces = flushTraces
	}

	// RegisterStatsExporter will register an exporter where we will collect our stats.
	// The error from the RegisterStatsExporter would be nil if the proper stats exporter
	// was specified by the user.
	cleanupStats := noop

	flushStats, err := observ.RegisterStatsExporter(r, conf.StatsExporter, Service)
	if err != nil {
		logger.Info(err)
	} else {
		cleanupStats = flushStats
	}

	var once sync.Once

	cleanup := func() {
		once.Do(func() {
			cleanupTraces()
			cleanupStats()
		})
	}

	user, pass, ok := conf.BasicAuth()
	if ok {
		r.Use(basicAuth(user, pass))
	}

	if !conf.FilterOff() {
		mf, err := module.NewFilter(conf.FilterFile)
		if err != nil {
			return nil, cleanup, fmt.Errorf("creating new filter: %w", err)
		}

		r.Use(mw.NewFilterMiddleware(mf, conf.GlobalEndpoint))
	}

	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	// Having the hook set means we want to use it
	if vHook := conf.ValidatorHook; vHook != "" {
		r.Use(mw.NewValidationMiddleware(client, vHook))
	}

	store, err := GetStorage(conf.StorageType, conf.Storage, conf.TimeoutDuration(), client)
	if err != nil {
		return nil, cleanup, fmt.Errorf("getting storage configuration: %w", err)
	}

	proxyRouter := r
	if subRouter != nil {
		proxyRouter = subRouter
	}

	err = addProxyRoutes(context.Background(), proxyRouter, store, logger, conf)
	if err != nil {
		return nil, cleanup, fmt.Errorf("adding proxy routes: %w", err)
	}

	h := otelhttp.NewHandler(r, "athens-proxy")

	return h, cleanup, nil
}
