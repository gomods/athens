package main

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// runMainEnv makes the test binary behave as the proxy itself: TestMain calls
// main() instead of running the tests. The end-to-end tests re-exec the test
// binary this way, so they exercise the real entry point without building a
// separate binary.
const runMainEnv = "ATHENS_E2E_RUN_MAIN"

const (
	startupTimeout = 15 * time.Second
	exitTimeout    = 15 * time.Second
	pollInterval   = 50 * time.Millisecond
)

func TestMain(m *testing.M) {
	if os.Getenv(runMainEnv) == "1" {
		main() // never returns: main exits the process
		os.Exit(2)
	}
	os.Exit(m.Run())
}

// syncBuffer is a bytes.Buffer safe for use while a process is still writing.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// proxyProc is a running proxy process started by startProxy.
type proxyProc struct {
	cmd    *exec.Cmd
	out    *syncBuffer
	done   chan struct{} // closed once the process has exited and been reaped
	waited error         // result of cmd.Wait; valid after done is closed
}

// startProxy re-executes the test binary as the proxy with the dev config and
// the given environment overrides. The process is killed when the test ends.
func startProxy(t *testing.T, env map[string]string) *proxyProc {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("end-to-end tests rely on POSIX signals")
	}

	cmd := exec.Command(os.Args[0], "-config_file=../../config.dev.toml") //nolint:gosec,noctx // re-exec of the test binary; killed via t.Cleanup.
	// Start from a clean environment so that settings of the machine running the
	// tests cannot leak into the proxy.
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "ATHENS_") || strings.HasPrefix(kv, "PORT=") {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, runMainEnv+"=1")
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	p := &proxyProc{cmd: cmd, out: &syncBuffer{}, done: make(chan struct{})}
	cmd.Stdout = p.out
	cmd.Stderr = p.out
	if err := cmd.Start(); err != nil {
		t.Fatalf("could not start proxy: %v", err)
	}
	go func() {
		p.waited = cmd.Wait()
		close(p.done)
	}()
	t.Cleanup(func() {
		_ = cmd.Process.Kill() // no-op error if it already exited
		<-p.done
		if t.Failed() {
			t.Logf("proxy output:\n%s", p.out.String())
		}
	})
	return p
}

// exitStatus waits for the process to exit and returns its exit status.
func (p *proxyProc) exitStatus(t *testing.T) int {
	t.Helper()
	select {
	case <-p.done:
	case <-time.After(exitTimeout):
		t.Fatalf("proxy did not exit within %v\noutput:\n%s", exitTimeout, p.out.String())
	}
	if p.waited == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if !errors.As(p.waited, &exitErr) {
		t.Fatalf("waiting for proxy: %v", p.waited)
	}
	return exitErr.ExitCode()
}

// waitForOutput blocks until the proxy output contains substr.
func (p *proxyProc) waitForOutput(t *testing.T, substr string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), startupTimeout)
	defer cancel()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for !strings.Contains(p.out.String(), substr) {
		select {
		case <-ctx.Done():
			t.Fatalf("output never contained %q\noutput:\n%s", substr, p.out.String())
		case <-p.done:
			if strings.Contains(p.out.String(), substr) {
				return
			}
			t.Fatalf("proxy exited before output contained %q\noutput:\n%s", substr, p.out.String())
		case <-ticker.C:
		}
	}
}

// waitHealthy polls /healthz until it answers 200.
func (p *proxyProc) waitHealthy(t *testing.T, addr string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), startupTimeout)
	defer cancel()
	client := &http.Client{Timeout: 2 * time.Second}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr+"/healthz", http.NoBody)
		if err != nil {
			t.Fatal(err)
		}
		if resp, err := client.Do(req); err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		select {
		case <-ctx.Done():
			t.Fatalf("proxy never became healthy on %s\noutput:\n%s", addr, p.out.String())
		case <-p.done:
			t.Fatalf("proxy exited before becoming healthy (%v)\noutput:\n%s", p.waited, p.out.String())
		case <-ticker.C:
		}
	}
}

func (p *proxyProc) terminate(t *testing.T) {
	t.Helper()
	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("could not signal proxy: %v", err)
	}
}

// freeAddr returns a loopback address that nothing is listening on.
func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}
	return addr
}

// busyAddr returns a loopback address that is held by a listener until the
// test ends.
func busyAddr(t *testing.T) string {
	t.Helper()
	ln, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	return ln.Addr().String()
}

func requireOutput(t *testing.T, p *proxyProc, substr string) {
	t.Helper()
	if !strings.Contains(p.out.String(), substr) {
		t.Fatalf("output does not contain %q\noutput:\n%s", substr, p.out.String())
	}
}

func requireExit(t *testing.T, p *proxyProc, want int) {
	t.Helper()
	if got := p.exitStatus(t); got != want {
		t.Fatalf("exit status = %d, want %d\noutput:\n%s", got, want, p.out.String())
	}
}

func TestE2EHealthyStartAndGracefulShutdown(t *testing.T) {
	addr := freeAddr(t)
	p := startProxy(t, map[string]string{
		"ATHENS_STORAGE_TYPE": "memory",
		"ATHENS_PORT":         addr,
	})

	p.waitHealthy(t, addr)
	p.terminate(t)

	requireExit(t, p, 0)
	requireOutput(t, p, "Shutting down server")
}

func TestE2EMainPortInUseIsFatal(t *testing.T) {
	p := startProxy(t, map[string]string{
		"ATHENS_STORAGE_TYPE": "memory",
		"ATHENS_PORT":         busyAddr(t),
	})

	requireExit(t, p, 1)
	requireOutput(t, p, "Could not listen")
}

func TestE2EPprofPortInUseDoesNotStopProxy(t *testing.T) {
	addr := freeAddr(t)
	p := startProxy(t, map[string]string{
		"ATHENS_STORAGE_TYPE": "memory",
		"ATHENS_PORT":         addr,
		"ATHENS_ENABLE_PPROF": "true",
		"ATHENS_PPROF_PORT":   busyAddr(t),
	})

	p.waitForOutput(t, "pprof server failed")
	p.waitHealthy(t, addr)
	select {
	case <-p.done:
		t.Fatalf("proxy exited after pprof failure (%v)\noutput:\n%s", p.waited, p.out.String())
	default:
	}

	p.terminate(t)
	requireExit(t, p, 0)
	requireOutput(t, p, "Shutting down server")
}

func TestE2EAppCreationFailureIsFatal(t *testing.T) {
	p := startProxy(t, map[string]string{
		"ATHENS_STORAGE_TYPE": "memory",
		"ATHENS_PORT":         freeAddr(t),
		// Passes config validation but App refuses to start with both set.
		"ATHENS_GITHUB_TOKEN": "unused",
		"ATHENS_NETRC_PATH":   "/nonexistent/netrc",
	})

	requireExit(t, p, 1)
	requireOutput(t, p, "Could not create App")
}
