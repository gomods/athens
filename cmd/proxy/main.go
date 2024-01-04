package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"time"

	"github.com/gomods/athens/cmd/proxy/actions"
	"github.com/gomods/athens/internal/shutdown"
	"github.com/gomods/athens/pkg/build"
	"github.com/gomods/athens/pkg/config"
	athenslog "github.com/gomods/athens/pkg/log"
	"github.com/sirupsen/logrus"
)

var (
	configFile = flag.String("config_file", "", "The path to the config file")
	version    = flag.Bool("version", false, "Print version information and exit")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Println(build.String())
		os.Exit(0)
	}
	conf, err := config.Load(*configFile)
	if err != nil {
		stdlog.Fatalf("Could not load config file: %v", err)
	}

	logLvl, err := logrus.ParseLevel(conf.LogLevel)
	if err != nil {
		stdlog.Fatalf("Could not parse log level %q: %v", conf.LogLevel, err)
	}

	logger := athenslog.New(conf.CloudRuntime, logLvl)

	handler, err := actions.App(logger, conf)
	if err != nil {
		logger.WithError(err).Fatal("Could not create App")
	}

	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 2 * time.Second,
	}
	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, shutdown.GetSignals()...)
		s := <-sigint
		logger.WithField("signal", s).Infof("Received signal, shutting down server")

		// We received an interrupt signal, shut down.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(conf.ShutdownTimeout))
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.WithError(err).Fatal("Could not shut down server")
		}
		close(idleConnsClosed)
	}()

	if conf.EnablePprof {
		go func() {
			// pprof to be exposed on a different port than the application for security matters,
			// not to expose profiling data and avoid DoS attacks (profiling slows down the service)
			// https://www.farsightsecurity.com/txt-record/2016/10/28/cmikk-go-remote-profiling/
			logger.WithField("port", conf.PprofPort).Infof("starting pprof")
			logger.Fatal(http.ListenAndServe(conf.PprofPort, nil)) //nolint:gosec // This should not be exposed to the world.
		}()
	}

	// Unix socket configuration, if available, takes precedence over TCP port configuration.
	var ln net.Listener

	if conf.UnixSocket != "" {
		logger := logger.WithField("unixSocket", conf.UnixSocket)
		logger.Info("Starting application")

		ln, err = net.Listen("unix", conf.UnixSocket)
		if err != nil {
			logger.WithError(err).Fatal("Could not listen on Unix domain socket")
		}
	} else {
		logger := logger.WithField("tcpPort", conf.Port)
		logger.Info("Starting application")

		ln, err = net.Listen("tcp", conf.Port)
		if err != nil {
			logger.WithError(err).Fatal("Could not listen on TCP port")
		}
	}

	if conf.TLSCertFile != "" && conf.TLSKeyFile != "" {
		err = srv.ServeTLS(ln, conf.TLSCertFile, conf.TLSKeyFile)
	} else {
		err = srv.Serve(ln)
	}

	if !errors.Is(err, http.ErrServerClosed) {
		logger.WithError(err).Fatal("Could not start server")
	}

	<-idleConnsClosed
}
