package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

type fakeLogger struct {
	infos []string
	errs  []string
}

func (l *fakeLogger) Infof(format string, args ...any) {
	l.infos = append(l.infos, fmt.Sprintf(format, args...))
}

func (l *fakeLogger) Errorf(format string, args ...any) {
	l.errs = append(l.errs, fmt.Sprintf(format, args...))
}

type fakeShutdowner struct {
	err            error
	deadline       time.Time
	hasDL          bool
	closed         bool
	ctxErr         error  // ctx.Err() as seen by Shutdown
	beforeShutdown func() // runs when Shutdown starts
}

func (f *fakeShutdowner) Shutdown(ctx context.Context) error {
	if f.beforeShutdown != nil {
		f.beforeShutdown()
	}
	f.ctxErr = ctx.Err()
	f.deadline, f.hasDL = ctx.Deadline()
	return f.err
}

func (f *fakeShutdowner) Close() error {
	f.closed = true
	return nil
}

func TestShutdownServerSuccess(t *testing.T) {
	f := &fakeShutdowner{}
	if err := shutdownServer(t.Context(), f, 5*time.Second); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if f.closed {
		t.Fatal("a clean shutdown must not force-close the server")
	}
	if !f.hasDL {
		t.Fatal("expected shutdown context to carry a deadline")
	}
	if remaining := time.Until(f.deadline); remaining <= 0 || remaining > 5*time.Second {
		t.Fatalf("deadline not derived from timeout, remaining=%v", remaining)
	}
}

func TestShutdownServerPropagatesError(t *testing.T) {
	want := context.DeadlineExceeded
	f := &fakeShutdowner{err: want}
	err := shutdownServer(t.Context(), f, time.Second)
	if err == nil {
		t.Fatal("expected shutdown failure to be returned, not swallowed")
	}
	if !f.closed {
		t.Fatal("expected the server to be force-closed when graceful shutdown fails")
	}
	if !errors.Is(err, want) {
		t.Fatalf("expected error to wrap %v, got %v", want, err)
	}
}

func TestWaitForShutdownSignal(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel() // the shutdown signal has arrived
	f := &fakeShutdowner{}
	log := &fakeLogger{}

	if got := waitForShutdown(ctx, func() {}, make(chan error), f, time.Second, log); got != 0 {
		t.Fatalf("exit code = %d, want 0", got)
	}
	if !f.hasDL {
		t.Fatal("expected the server to be shut down")
	}
	if len(log.errs) != 0 {
		t.Fatalf("unexpected errors logged: %v", log.errs)
	}
}

// A failure of the main listener must stop the proxy
// with a non-zero code, but by returning so the caller's cleanup still runs,
// not by exiting on the spot.
func TestWaitForShutdownFatalError(t *testing.T) {
	fatal := make(chan error, 1)
	fatal <- errors.New("server failed: boom")
	f := &fakeShutdowner{}
	log := &fakeLogger{}

	// The signal context is never cancelled, so returning proves the fatal
	// error ended the wait.
	done := make(chan int, 1)
	go func() { done <- waitForShutdown(t.Context(), func() {}, fatal, f, time.Second, log) }()

	select {
	case got := <-done:
		if got != 1 {
			t.Fatalf("exit code = %d, want 1", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("waitForShutdown did not return after a fatal error")
	}
	if !f.hasDL {
		t.Fatal("expected the server to be shut down after a fatal error")
	}
	if len(log.errs) != 1 || !strings.Contains(log.errs[0], "boom") {
		t.Fatalf("expected the fatal error to be logged, got %v", log.errs)
	}
}

func TestWaitForShutdownShutdownFails(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	f := &fakeShutdowner{err: context.DeadlineExceeded}
	log := &fakeLogger{}

	if got := waitForShutdown(ctx, func() {}, make(chan error), f, time.Second, log); got != 1 {
		t.Fatalf("exit code = %d, want 1", got)
	}
	if !f.closed {
		t.Fatal("expected the server to be force-closed")
	}
	if len(log.errs) != 1 {
		t.Fatalf("expected the shutdown failure to be logged once, got %v", log.errs)
	}
}

// Once shutdown starts, a second SIGINT/SIGTERM must get the default behaviour
// (terminate) instead of being swallowed while in-flight requests drain, so the
// signal handler has to be released before the shutdown begins.
func TestWaitForShutdownReleasesSignalsBeforeShuttingDown(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	stopped := false
	var stoppedBeforeShutdown bool
	f := &fakeShutdowner{beforeShutdown: func() { stoppedBeforeShutdown = stopped }}

	waitForShutdown(ctx, func() { stopped = true }, make(chan error), f, time.Second, &fakeLogger{})

	if !stopped {
		t.Fatal("expected the signal handler to be released")
	}
	if !stoppedBeforeShutdown {
		t.Fatal("expected the signal handler to be released before Shutdown starts")
	}
}

// The signal context is already cancelled by the time we shut down, and must
// not cut the graceful shutdown short.
func TestShutdownServerIgnoresCancelledParent(t *testing.T) {
	parent, cancel := context.WithCancel(t.Context())
	cancel()
	f := &fakeShutdowner{}

	if err := shutdownServer(parent, f, 5*time.Second); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.ctxErr != nil {
		t.Fatalf("Shutdown got an already-done context: %v", f.ctxErr)
	}
	if !f.hasDL {
		t.Fatal("expected the shutdown context to carry the timeout")
	}
}

func TestExitCode(t *testing.T) {
	if got := exitCode(nil); got != 0 {
		t.Fatalf("exitCode(nil) = %d, want 0", got)
	}
	if got := exitCode(errors.New("boom")); got != 1 {
		t.Fatalf("exitCode(err) = %d, want 1", got)
	}
}

func TestNewServerTimeouts(t *testing.T) {
	srv := newServer(http.NewServeMux())
	if srv.IdleTimeout != idleTimeout || idleTimeout <= 0 {
		t.Fatalf("IdleTimeout = %v, want positive %v", srv.IdleTimeout, idleTimeout)
	}
	if srv.ReadHeaderTimeout <= 0 {
		t.Fatal("ReadHeaderTimeout must stay set")
	}
	// Athens streams large module zips and runs long synchronous downloads,
	// so whole-request read/write timeouts would break them.
	if srv.ReadTimeout != 0 || srv.WriteTimeout != 0 {
		t.Fatalf("ReadTimeout/WriteTimeout must stay unset, got %v/%v", srv.ReadTimeout, srv.WriteTimeout)
	}
}
