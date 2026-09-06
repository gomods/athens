//go:build unix

package shutdown

import (
	"errors"
	"os/exec"
	"strings"
	"syscall"
	"testing"
)

// TestExternalReaperStealsExecExitCode documents gomods/athens#2049.
//
// v0.16.0 added an in-process reaper that listened for SIGCHLD and called
// syscall.Wait4(-1, ...) ("reap ANY child"). That races os/exec, which waits on
// its own specific child to collect the exit code. When the reaper wins,
// os/exec's wait returns ECHILD and Cmd.Wait() fails with "no child processes"
// instead of the real exit status, which surfaced as spurious 404s.
//
// This test reproduces the failure mode deterministically (no reliance on the
// race firing): it reaps a child out from under os/exec and shows that Wait()
// then loses the exit code. It is here to document why Athens must NOT run an
// in-process Wait4(-1) reaper, and instead relies on an init (tini) at PID 1 to
// reap orphaned subprocesses. See internal/shutdown and the Docker install docs.
func TestExternalReaperStealsExecExitCode(t *testing.T) {
	truePath, err := exec.LookPath("true")
	if err != nil {
		t.Skipf("`true` binary not found: %v", err)
	}

	cmd := exec.Command(truePath)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start command: %v", err)
	}

	// Stand in for a Wait4(-1) reaper that wins the race: reap the child before
	// os/exec gets to it. Wait4 blocks (options 0) until the child exits.
	var wstatus syscall.WaitStatus
	reaped, err := syscall.Wait4(-1, &wstatus, 0, nil)
	if err != nil {
		t.Fatalf("reaper Wait4 failed: %v", err)
	}
	if reaped != cmd.Process.Pid {
		t.Fatalf("reaped pid %d, expected the command's pid %d", reaped, cmd.Process.Pid)
	}

	// os/exec can no longer collect the exit status; its wait gets ECHILD.
	waitErr := cmd.Wait()
	if waitErr == nil {
		t.Fatal("expected Cmd.Wait() to fail after the child was reaped externally, got nil")
	}
	if !errors.Is(waitErr, syscall.ECHILD) && !strings.Contains(waitErr.Error(), "no child processes") {
		t.Fatalf("expected an ECHILD / \"no child processes\" error, got: %v", waitErr)
	}

	t.Logf("confirmed #2049: an external reaper took exit status %d, and os/exec.Cmd.Wait() returned %q",
		wstatus.ExitStatus(), waitErr)
}
