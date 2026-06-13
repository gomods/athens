package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	stdlog "log"
	"log/slog"
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

	logLvl, err := athenslog.ParseLevel(conf.LogLevel)
	if err != nil {
		stdlog.Fatalf("Could not parse log level %q: %v", conf.LogLevel, err)
	}

	logger := athenslog.New(conf.CloudRuntime, logLvl, conf.LogFormat)

	// Route the standard library logger's output through our logger at the
	// error level.
	stdlog.SetOutput(logger.StdLogger(slog.LevelError).Writer())
	stdlog.SetFlags(stdlog.Flags() &^ (stdlog.Ldate | stdlog.Ltime))

	handler, cleanup, err := actions.App(logger, conf)
	if err != nil {
		logger.Fatalf("Could not create App: %v", err)
	}
	defer cleanup()

	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 2 * time.Second,
	}

	if conf.EnablePprof {
		go func() {
			// pprof to be exposed on a different port than the application for security matters,
			// not to expose profiling data and avoid DoS attacks (profiling slows down the service)
			// https://www.farsightsecurity.com/txt-record/2016/10/28/cmikk-go-remote-profiling/
			logger.WithFields(map[string]any{"port": conf.PprofPort}).Infof("starting pprof")
			logger.Fatalf("pprof server failed: %v", http.ListenAndServe(conf.PprofPort, nil)) //nolint:gosec // This should not be exposed to the world.
		}()
	}

	// Unix socket configuration, if available, takes precedence over TCP port configuration.
	var ln net.Listener

	if conf.UnixSocket != "" {
		logger.WithFields(map[string]any{"unixSocket": conf.UnixSocket}).Infof("Starting application")

		//nolint:noctx
		ln, err = net.Listen("unix", conf.UnixSocket)
		if err != nil {
			logger.Fatalf("Could not listen on Unix domain socket %q: %v", conf.UnixSocket, err)
		}
	} else {
		logger.WithFields(map[string]any{"tcpPort": conf.Port}).Infof("Starting application")

		//nolint:noctx
		ln, err = net.Listen("tcp", conf.Port)
		if err != nil {
			logger.Fatalf("Could not listen on TCP port %q: %v", conf.Port, err)
		}
	}

	signalCtx, signalStop := signal.NotifyContext(context.Background(), shutdown.GetSignals()...)
	reaper := shutdown.ChildProcReaper(signalCtx, logger)

	go func() {
		defer signalStop()
		if conf.TLSCertFile != "" && conf.TLSKeyFile != "" {
			err = srv.ServeTLS(ln, conf.TLSCertFile, conf.TLSKeyFile)
		} else {
			err = srv.Serve(ln)
		}

		if !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Could not start server: %v", err)
		}
	}()

	// Wait for shutdown signal, then cleanup before exit.
	<-signalCtx.Done()
	logger.Infof("Shutting down server")

	// We received an interrupt signal, shut down.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(conf.ShutdownTimeout))
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("Could not shut down server: %v", err)
	}
	<-reaper.Done()
}
