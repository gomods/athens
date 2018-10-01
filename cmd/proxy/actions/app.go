package actions

import (
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/csrf"
	"github.com/gobuffalo/buffalo/middleware/ssl"
	"github.com/gobuffalo/buffalo/render"
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
	store, err := GetStorage(conf.Proxy.StorageType, conf.Storage)
	if err != nil {
		err = fmt.Errorf("error getting storage configuration (%s)", err)
		return nil, err
	}

	// mount .netrc to home dir
	// to have access to private repos.
	initializeNETRC(conf.Proxy.NETRCPath)

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
		Addr:        conf.Proxy.Port,
		WorkerOff:   true,
		Host:        "http://127.0.0.1" + conf.Proxy.Port,
	})
	if prefix := conf.Proxy.PathPrefix; prefix != "" {
		// certain Ingress Controllers (such as GCP Load Balancer)
		// can not send custom headers and therefore if the proxy
		// is running behind a prefix as well as some authentication
		// mechanism, we should allow the plain / to return 200.
		app.GET("/", healthHandler)
		app = app.Group(prefix)
	}

	// Register exporter to export traces
	exporter, err := observ.RegisterTraceExporter(conf.TraceExporterURL, Service, ENV)
	if err != nil {
		lggr.Infof("%s", err)
	} else {
		defer exporter.Flush()
		app.Use(observ.Tracer(Service))
	}

	// Automatically redirect to SSL
	app.Use(ssl.ForceSSL(secure.Options{
		SSLRedirect:     conf.Proxy.ForceSSL,
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	}))

	if ENV == "development" {
		app.Use(middleware.ParameterLogger)
	}

	initializeAuth(app)
	// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
	// Remove to disable this.
	if conf.EnableCSRFProtection {
		csrfMiddleware := csrf.New
		app.Use(csrfMiddleware)
	}

	if !conf.Proxy.FilterOff {
		mf := module.NewFilter(conf.FilterFile)
		app.Use(mw.NewFilterMiddleware(mf, conf.Proxy.OlympusGlobalEndpoint))
	}

	// Having the hook set means we want to use it
	if vHook := conf.Proxy.ValidatorHook; vHook != "" {
		app.Use(mw.LogEntryMiddleware(mw.NewValidationMiddleware, lggr, vHook))
	}

	user, pass, ok := conf.Proxy.BasicAuth()
	if ok {
		app.Use(basicAuth(user, pass))
	}

	if err := addProxyRoutes(app, store, lggr, conf.GoBinary, conf.GoGetWorkers, conf.ProtocolWorkers); err != nil {
		err = fmt.Errorf("error adding proxy routes (%s)", err)
		return nil, err
	}

	return app, nil
}
