package actions

import (
	"fmt"
	"net/http"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/log"
	mw "github.com/gomods/athens/pkg/middleware"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gorilla/mux"
	"github.com/unrolled/secure"
	"go.opencensus.io/plugin/ochttp"
)

// Service is the name of the service that we want to tag our processes with.
const Service = "proxy"

// App is where all routes and middleware for the proxy
// should be defined. This is the nerve center of your
// application.
func App(logger *log.Logger, conf *config.Config) (http.Handler, error) {
	if conf.GithubToken != "" {
		if conf.NETRCPath != "" {
			return nil, fmt.Errorf("cannot provide both GithubToken and NETRCPath")
		}

		if err := netrcFromToken(conf.GithubToken); err != nil {
			return nil, fmt.Errorf("creating netrc from token: %w", err)
		}
	}

	// mount .netrc to home dir
	// to have access to private repos.
	if err := initializeAuthFile(conf.NETRCPath); err != nil {
		return nil, fmt.Errorf("initializing auth file from netrc: %w", err)
	}

	// mount .hgrc to home dir
	// to have access to private repos.
	if err := initializeAuthFile(conf.HGRCPath); err != nil {
		return nil, fmt.Errorf("initializing auth file from hgrc: %w", err)
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
	// RegisterExporter returns the function that all traces are flushed to the exporter
	// and the exporter needs to be stopped. The function should be called when the exporter
	// is no longer needed.
	flushTraces, err := observ.RegisterExporter(
		conf.TraceExporter,
		conf.TraceExporterURL,
		Service,
		conf.GoEnv,
	)
	if err != nil {
		logger.Info(err)
	} else {
		defer flushTraces()
	}

	// RegisterStatsExporter will register an exporter where we will collect our stats.
	// The error from the RegisterStatsExporter would be nil if the proper stats exporter
	// was specified by the user.
	flushStats, err := observ.RegisterStatsExporter(r, conf.StatsExporter, Service)
	if err != nil {
		logger.Info(err)
	} else {
		defer flushStats()
	}

	user, pass, ok := conf.BasicAuth()
	if ok {
		r.Use(basicAuth(user, pass))
	}

	if !conf.FilterOff() {
		mf, err := module.NewFilter(conf.FilterFile)
		if err != nil {
			return nil, fmt.Errorf("creating new filter: %w", err)
		}
		r.Use(mw.NewFilterMiddleware(mf, conf.GlobalEndpoint))
	}

	client := &http.Client{
		Transport: &ochttp.Transport{
			Base: http.DefaultTransport,
		},
	}

	// Having the hook set means we want to use it
	if vHook := conf.ValidatorHook; vHook != "" {
		r.Use(mw.NewValidationMiddleware(client, vHook))
	}

	store, err := GetStorage(conf.StorageType, conf.Storage, conf.TimeoutDuration(), client)
	if err != nil {
		return nil, fmt.Errorf("getting storage configuration: %w", err)
	}

	proxyRouter := r
	if subRouter != nil {
		proxyRouter = subRouter
	}
	if err := addProxyRoutes(proxyRouter, store, logger, conf); err != nil {
		return nil, fmt.Errorf("adding proxy routes: %w", err)
	}

	h := &ochttp.Handler{
		Handler: r,
	}

	return h, nil
}
