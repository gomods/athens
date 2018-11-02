package actions

import (
	"fmt"
	"os"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	forcessl "github.com/gobuffalo/mw-forcessl"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/log"
	mw "github.com/gomods/athens/pkg/middleware"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/observ"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/unrolled/secure"
)

// Service is the name of the service that we want to tag our processes with
const Service = "proxy"

var proxy = render.New(render.Options{
	// Add template helpers here:
	Helpers: render.Helpers{},
})

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App(conf *config.Config) (*buffalo.App, error) {
	// ENV is used to help switch settings based on where the
	// application is being run. Default is "development".
	ENV := conf.GoEnv
	store, err := GetStorage(conf.StorageType, conf.Storage)
	if err != nil {
		err = fmt.Errorf("error getting storage configuration (%s)", err)
		return nil, err
	}

	if conf.GithubToken != "" {
		if conf.NETRCPath != "" {
			fmt.Println("Cannot provide both GithubToken and NETRCPath. Only provide one.")
			os.Exit(1)
		}

		netrcFromToken(conf.GithubToken)
	}

	// mount .netrc to home dir
	// to have access to private repos.
	initializeAuthFile(conf.NETRCPath)

	// mount .hgrc to home dir
	// to have access to private repos.
	initializeAuthFile(conf.HGRCPath)

	logLvl, err := logrus.ParseLevel(conf.LogLevel)
	if err != nil {
		return nil, err
	}
	lggr := log.New(conf.CloudRuntime, logLvl)

	bLogLvl, err := logrus.ParseLevel(conf.BuffaloLogLevel)
	if err != nil {
		return nil, err
	}
	blggr := log.Buffalo(bLogLvl)

	app := buffalo.New(buffalo.Options{
		Env: ENV,
		PreWares: []buffalo.PreWare{
			cors.Default().Handler,
		},
		SessionName: "_athens_session",
		Logger:      blggr,
		Addr:        conf.Port,
		WorkerOff:   true,
		Host:        "http://127.0.0.1" + conf.Port,
	})

	app.Use(mw.LogEntryMiddleware(lggr))

	if prefix := conf.PathPrefix; prefix != "" {
		// certain Ingress Controllers (such as GCP Load Balancer)
		// can not send custom headers and therefore if the proxy
		// is running behind a prefix as well as some authentication
		// mechanism, we should allow the plain / to return 200.
		app.GET("/", healthHandler)
		app = app.Group(prefix)
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
		ENV,
	)
	if err != nil {
		lggr.Infof("%s", err)
	} else {
		defer flushTraces()
		app.Use(observ.Tracer(Service))
	}

	// Automatically redirect to SSL
	app.Use(forcessl.Middleware(secure.Options{
		SSLRedirect:     conf.ForceSSL,
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	}))

	if ENV == "development" {
		app.Use(paramlogger.ParameterLogger)
	}

	initializeAuth(app)

	if !conf.FilterOff() {
		mf, err := module.NewFilter(conf.FilterFile)
		if err != nil {
			lggr.Fatal(err)
		}
		app.Use(mw.NewFilterMiddleware(mf, conf.GlobalEndpoint))
	}

	// Having the hook set means we want to use it
	if vHook := conf.ValidatorHook; vHook != "" {
		app.Use(mw.NewValidationMiddleware(vHook))
	}

	user, pass, ok := conf.BasicAuth()
	if ok {
		app.Use(basicAuth(user, pass))
	}

	if err := addProxyRoutes(app, store, lggr, conf.GoBinary, conf.GoGetWorkers, conf.ProtocolWorkers); err != nil {
		err = fmt.Errorf("error adding proxy routes (%s)", err)
		return nil, err
	}

	return app, nil
}
