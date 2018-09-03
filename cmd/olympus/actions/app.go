package actions

import (
	"fmt"
	stdlog "log"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/csrf"
	"github.com/gobuffalo/buffalo/middleware/i18n"
	"github.com/gobuffalo/buffalo/middleware/ssl"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/gocraft-work-adapter"
	"github.com/gobuffalo/packr"
	"github.com/gocraft/work"
	"github.com/gomods/athens/pkg/cdn/metadata/azurecdn"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/eventlog"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/stash"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomodule/redigo/redis"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/unrolled/secure"
)

type workerConfig struct {
	store           storage.Backend
	eLog            eventlog.Eventlog
	wType           string
	redisEndpoint   string
	maxConc         int
	maxFails        uint
	downloadTimeout time.Duration
}

const (
	// OlympusWorkerName is the name of the Olympus worker
	OlympusWorkerName = "olympus-worker"
	// DownloadHandlerName is name of the handler downloading packages from VCS
	DownloadHandlerName = "download-handler"
	// PushNotificationHandlerName is the name of the handler processing push notifications
	PushNotificationHandlerName = "push-notification-worker"
)

var (
	workerQueue               = "default"
	workerModuleKey           = "module"
	workerVersionKey          = "version"
	workerPushNotificationKey = "push-notification"
	// T is buffalo Translator
	T *i18n.Translator
)

// App is where all routes and middleware for buffalo should be defined.
// This is the nerve center of your application.
func App(conf *config.Config) (*buffalo.App, error) {
	// ENV is used to help switch settings based on where the
	// application is being run. Default is "development".
	ENV := conf.GoEnv
	port := conf.Olympus.Port

	storage, err := GetStorage(conf.Olympus.StorageType, conf.Storage)
	if err != nil {
		return nil, err
	}
	if conf.Storage == nil || conf.Storage.Mongo == nil {
		return nil, fmt.Errorf("A valid Mongo configuration is required to create the event log")
	}
	eLog, err := GetEventLog(conf.Storage.Mongo.URL, conf.Storage.Mongo.CertPath)
	if err != nil {
		return nil, fmt.Errorf("error creating eventlog (%s)", err)
	}
	wConf := workerConfig{
		store:           storage,
		eLog:            eLog,
		wType:           conf.Olympus.WorkerType,
		maxConc:         conf.MaxConcurrency,
		maxFails:        conf.MaxWorkerFails,
		downloadTimeout: conf.TimeoutDuration(),
		redisEndpoint:   conf.Olympus.RedisQueueAddress,
	}
	w, err := getWorker(wConf)
	if err != nil {
		return nil, err
	}

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
		Addr: port,
		Env:  ENV,
		PreWares: []buffalo.PreWare{
			cors.Default().Handler,
		},
		SessionName: "_olympus_session",
		Worker:      w,
		WorkerOff:   true, // TODO(marwan): turned off until worker is being used.
		Logger:      blggr,
	})

	// Automatically redirect to SSL
	app.Use(ssl.ForceSSL(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	}))

	if ENV == "development" {
		app.Use(middleware.ParameterLogger)
	}
	initializeTracing(app)
	// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
	// Remove to disable this.
	if conf.EnableCSRFProtection {
		csrfMiddleware := csrf.New
		app.Use(csrfMiddleware)
	}

	// TODO: parameterize the GoGet getter here.
	//
	// Defaulting to Azure for now
	app.Use(GoGet(azurecdn.Metadata{
		// TODO: initialize the azurecdn.Storage struct here
	}))

	// Setup and use translations:
	if T, err = i18n.New(packr.NewBox("../locales"), "en-US"); err != nil {
		app.Stop(err)
	}
	app.Use(T.Middleware())

	app.GET("/diff/{lastID}", diffHandler(storage, eLog))
	app.GET("/feed/{lastID}", feedHandler(storage))
	app.GET("/eventlog/{sequence_id}", eventlogHandler(eLog))
	app.POST("/cachemiss", cachemissHandler(w))
	app.POST("/push", pushNotificationHandler(w))
	app.GET("/healthz", healthHandler)

	// Download Protocol
	goBin := conf.GoBinary
	fs := afero.NewOsFs()
	mf, err := module.NewGoGetFetcher(goBin, fs)
	if err != nil {
		return nil, err
	}
	st := stash.New(mf, storage)
	dpOpts := &download.Opts{
		Storage:   storage,
		Stasher:   st,
		GoBinPath: goBin,
		Fs:        fs,
	}
	dp := download.New(dpOpts)

	handlerOpts := &download.HandlerOpts{Protocol: dp, Logger: lggr, Engine: renderEng}
	download.RegisterHandlers(app, handlerOpts)

	app.ServeFiles("/", assetsBox) // serve files from the public directory

	return app, nil
}

func getWorker(wConf workerConfig) (worker.Worker, error) {
	switch wConf.wType {
	case "redis":
		return registerRedis(wConf)
	case "memory":
		return registerInMem(wConf)
	default:
		stdlog.Printf("Provided background worker type %s. Expected redis|memory. Defaulting to memory", wConf.wType)
		return registerInMem(wConf)
	}
}

func registerInMem(wConf workerConfig) (worker.Worker, error) {
	w := worker.NewSimple()
	if err := w.Register(PushNotificationHandlerName, GetProcessPushNotificationJob(wConf.store, wConf.eLog, wConf.downloadTimeout)); err != nil {
		return nil, err
	}
	return w, nil
}

func registerRedis(wConf workerConfig) (worker.Worker, error) {
	port := wConf.redisEndpoint
	w := gwa.New(gwa.Options{
		Pool: &redis.Pool{
			MaxActive: 5,
			MaxIdle:   5,
			Wait:      true,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", port)
			},
		},
		Name:           OlympusWorkerName,
		MaxConcurrency: wConf.maxConc,
	})

	opts := work.JobOptions{
		SkipDead: true,
		MaxFails: wConf.maxFails,
	}

	return w, w.RegisterWithOptions(PushNotificationHandlerName, opts, GetProcessPushNotificationJob(wConf.store, wConf.eLog, wConf.downloadTimeout))
}
