package actions

import (
	"log"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/ssl"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/envy"
	"github.com/rs/cors"
	"github.com/unrolled/secure"

	"github.com/gobuffalo/buffalo/middleware/csrf"
	"github.com/gobuffalo/buffalo/middleware/i18n"
	"github.com/gobuffalo/packr"
)

const (
	// WorkerName is the name of the worker processing misses
	WorkerName                = "pushProcessor"
	workerQueue               = "default"
	workerPushNotificationKey = "pushNotification"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

// T is buffalo Translator
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		redisPort := envy.Get("ATHENS_REDIS_QUEUE_PORT", ":6379")
		app = buffalo.New(buffalo.Options{
			Env: ENV,
			PreWares: []buffalo.PreWare{
				cors.Default().Handler,
			},
			SessionName: "_olympus_session",
			Worker:      getWorker(redisPort),
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
		csrfMiddleware := csrf.New
		app.Use(csrfMiddleware)

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.PopTransaction)
		// Remove to disable this.
		// app.Use(middleware.PopTransaction(models.DB))

		// Setup and use translations:
		var err error
		if T, err = i18n.New(packr.NewBox("../locales"), "en-US"); err != nil {
			app.Stop(err)
		}
		app.Use(T.Middleware())

		storage, err := newStorage()
		if err != nil {
			log.Fatalf("error creating storage (%s)", err)
			return nil
		}
		eventlogReader, err := newEventlog()
		if err != nil {
			log.Fatalf("error creating eventlog (%s)", err)
			return nil
		}

		cacheMissesLog, err := newCacheMissesLog()
		if err != nil {
			log.Fatalf("error creating cachemisses log (%s)", err)
			return nil
		}

		app.GET("/", homeHandler)
		app.GET("/feed/{syncpoint:.*}", feedHandler(storage))
		app.GET("/eventlog/{sequence_id}", eventlogHandler(eventlogReader))
		app.POST("/cachemiss", cachemissHandler(cacheMissesLog))
		app.ServeFiles("/", assetsBox) // serve files from the public directory
	}

	return app
}

func getWorker(port string) worker.Worker {
	return gwa.New(gwa.Options{
		Pool: &redis.Pool{
			MaxActive: 5,
			MaxIdle:   5,
			Wait:      true,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", port)
			},
		},
		Name:           WorkerName,
		MaxConcurrency: 25,
	})
}
