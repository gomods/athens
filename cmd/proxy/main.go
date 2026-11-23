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
	configFile    = flag.String("config_file", "", "The path to the config file")
	version       = flag.Bool("version", false, "Print version information and exit")
	verifyStorage = flag.Bool("verify-storage", false, "Verify stored module zips against the checksum database and report mismatches, then exit (does not start the server)")
	purge         = flag.Bool("purge", false, "With -verify-storage, delete the mismatched module versions (requires -verify-storage)")
)

// idleTimeout bounds how long an idle keep-alive connection is kept open.
//
// There is deliberately no ReadTimeout or WriteTimeout: Athens streams large
// module zips and serves long synchronous downloads, and those server-wide
// whole-request timeouts would cut them off.
const idleTimeout = 2 * time.Minute

// shutdowner is the part of *http.Server needed to shut it down.
type shutdowner interface {
	Shutdown(ctx context.Context) error
	Close() error
}

func newServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       idleTimeout,
	}
}

// shutdownServer gracefully shuts srv down, giving in-flight requests up to
// timeout to finish. A failure (such as the timeout elapsing) is returned, and
// any connections still open are closed so that no handler is left running
// when the caller goes on to clean up.
func shutdownServer(parent context.Context, srv shutdowner, timeout time.Duration) error {
	// The parent is usually the already-cancelled signal context, so detach from
	// its cancellation and deadline while keeping its values.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), timeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		_ = srv.Close()
		return fmt.Errorf("could not shut down server: %w", err)
	}
	return nil
}

func exitCode(err error) int {
	if err != nil {
		return 1
	}
	return 0
}

// shutdownLogger is the part of the application logger waitForShutdown needs.
type shutdownLogger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
}

// waitForShutdown blocks until a shutdown signal arrives on signalCtx or a
// background component reports a failure on fatal, then shuts srv down. It
// returns the process exit code: non-zero if a component failed or the
// shutdown itself failed. Either way the server has been shut down by the time
// it returns, so the caller's deferred cleanup runs with no handler in flight.
// stopSignals is called as soon as the wait ends, before the shutdown begins.
func waitForShutdown(signalCtx context.Context, stopSignals func(), fatal <-chan error, srv shutdowner, timeout time.Duration, log shutdownLogger) int {
	var fatalErr error
	select {
	case <-signalCtx.Done():
	case fatalErr = <-fatal:
		log.Errorf("%v", fatalErr)
	}

	// Release the signal handler now so that a second SIGINT/SIGTERM terminates
	// the process as usual instead of being swallowed while requests drain.
	stopSignals()

	log.Infof("Shutting down server")

	shutdownErr := shutdownServer(signalCtx, srv, timeout)
	if shutdownErr != nil {
		log.Errorf("%v", shutdownErr)
	}
	return exitCode(errors.Join(fatalErr, shutdownErr))
}

// main only turns run's result into an exit status, so that the deferred
// cleanup in run always executes before the process exits.
func main() {
	os.Exit(run())
}

func run() int {
	flag.Parse()
	if err := actions.ValidateVerifyFlags(*verifyStorage, *purge); err != nil {
		stdlog.Fatalf("%v", err)
	}
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

	// Athens shells out to the Go toolchain (go, git, ssh) and relies on an init
	// at PID 1 to reap the orphaned subprocesses they leave behind. Running as
	// PID 1 means there is no init to do that, so warn the operator.
	if os.Getpid() == 1 {
		logger.Warnf("Athens is running as PID 1 with no init to reap subprocesses; " +
			"run it under an init such as tini or `docker/podman run --init` to avoid zombie processes")
	}

	if *verifyStorage {
		if err := actions.RunVerify(conf, *purge, os.Stdout); err != nil {
			logger.Errorf("verify-storage failed: %v", err)
			return 1
		}
		return 0
	}

	// App returns a usable cleanup func even when it fails, so defer it before
	// checking the error: a failed start must still flush traces and stats.
	handler, cleanup, err := actions.App(logger, conf)
	defer cleanup()
	if err != nil {
		logger.Errorf("Could not create App: %v", err)
		return 1
	}

	srv := newServer(handler)

	// fatal collects the failure of the main listener, which must stop the proxy.
	// Buffered so the sender never blocks, whatever state the main flow is in.
	fatal := make(chan error, 1)

	if conf.EnablePprof {
		go func() {
			// pprof to be exposed on a different port than the application for security matters,
			// not to expose profiling data and avoid DoS attacks (profiling slows down the service)
			// https://www.farsightsecurity.com/txt-record/2016/10/28/cmikk-go-remote-profiling/
			logger.WithFields(map[string]any{"port": conf.PprofPort}).Infof("starting pprof")
			// pprof is optional diagnostics: its failure is logged but must not take
			// the proxy down. ListenAndServe only ever returns a non-nil error.
			err := http.ListenAndServe(conf.PprofPort, nil) //nolint:gosec // This should not be exposed to the world.
			logger.Errorf("pprof server failed: %v", err)
		}()
	}

	// Unix socket configuration, if available, takes precedence over TCP port configuration.
	var ln net.Listener

	if conf.UnixSocket != "" {
		logger.WithFields(map[string]any{"unixSocket": conf.UnixSocket}).Infof("Starting application")

		//nolint:noctx
		ln, err = net.Listen("unix", conf.UnixSocket)
		if err != nil {
			logger.Errorf("Could not listen on Unix domain socket %q: %v", conf.UnixSocket, err)
			return 1
		}
	} else {
		logger.WithFields(map[string]any{"tcpPort": conf.Port}).Infof("Starting application")

		//nolint:noctx
		ln, err = net.Listen("tcp", conf.Port)
		if err != nil {
			logger.Errorf("Could not listen on TCP port %q: %v", conf.Port, err)
			return 1
		}
	}

	signalCtx, signalStop := signal.NotifyContext(context.Background(), shutdown.GetSignals()...)
	defer signalStop()

	go func() {
		var serveErr error
		if conf.TLSCertFile != "" && conf.TLSKeyFile != "" {
			serveErr = srv.ServeTLS(ln, conf.TLSCertFile, conf.TLSKeyFile)
		} else {
			serveErr = srv.Serve(ln)
		}

		if !errors.Is(serveErr, http.ErrServerClosed) {
			fatal <- fmt.Errorf("server failed: %w", serveErr)
		}
	}()

	return waitForShutdown(signalCtx, signalStop, fatal, srv, time.Second*time.Duration(conf.ShutdownTimeout), logger)
}
