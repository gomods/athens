package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "net/http/pprof"

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
		stdlog.Fatalf("could not load config file: %v", err)
	}

	logLvl, err := logrus.ParseLevel(conf.LogLevel)
	if err != nil {
		stdlog.Fatalf("failed logrus.ParseLevel(%q): %v", conf.LogLevel, err)
	}

	logger := athenslog.New(conf.CloudRuntime, logLvl)

	handler, err := actions.App(logger, conf)
	if err != nil {
		logger.WithError(err).Fatal("failed to create App")
	}

	cert, key, err := conf.TLSCertFiles()
	if err != nil {
		logger.WithError(err).Fatal("failed conf.TLSCertFiles")
	}

	srv := &http.Server{
		Addr:    conf.Port,
		Handler: handler,
	}
	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, shutdown.GetSignals()...)
		s := <-sigint
		logger.WithField("signal", s).Infof("received signal, shutting down server")

		// We received an interrupt signal, shut down.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(conf.ShutdownTimeout))
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.WithError(err).Fatal("failed srv.Shutdown")
		}
		close(idleConnsClosed)
	}()

	if conf.EnablePprof {
		go func() {
			// pprof to be exposed on a different port than the application for security matters,
			// not to expose profiling data and avoid DoS attacks (profiling slows down the service)
			// https://www.farsightsecurity.com/txt-record/2016/10/28/cmikk-go-remote-profiling/
			logger.WithField("port", conf.PprofPort).Infof("starting pprof")
			logger.Fatal(http.ListenAndServe(conf.PprofPort, nil))
		}()
	}

	logger.WithField("port", conf.Port).Infof("starting application")
	if cert != "" && key != "" {
		err = srv.ListenAndServeTLS(conf.TLSCertFile, conf.TLSKeyFile)
	} else {
		err = srv.ListenAndServe()
	}

	if !errors.Is(err, http.ErrServerClosed) {
		logger.WithError(err).Fatal("application error")
	}

	<-idleConnsClosed
}
