package stash

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestPoolWrapper(t *testing.T) {
	m := &mockPoolStasher{inputMod: "mod", inputVer: "ver", err: fmt.Errorf("wrapped err")}
	s := WithPool(2)(m)
	_, err := s.Stash(t.Context(), m.inputMod, m.inputVer)
	if err.Error() != m.err.Error() {
		t.Fatalf("expected err to be `%v` but got `%v`", m.err, err)
	}
}

type mockPoolStasher struct {
	inputMod string
	inputVer string
	err      error
}

func (m *mockPoolStasher) Stash(ctx context.Context, mod, ver string) (string, error) {
	if m.inputMod != mod {
		return "", fmt.Errorf("expected input mod %v but got %v", m.inputMod, mod)
	}
	if m.inputVer != ver {
		return "", fmt.Errorf("expected input ver %v but got %v", m.inputVer, ver)
	}
	return "", m.err
}

// blockingStasher blocks every Stash call until release is closed.
type blockingStasher struct {
	started chan struct{}
	release chan struct{}
}

func (b *blockingStasher) Stash(ctx context.Context, mod, ver string) (string, error) {
	b.started <- struct{}{}
	<-b.release
	return "newver", nil
}

type stashResult struct {
	ver string
	err error
}

func stashAsync(ctx context.Context, s Stasher) <-chan stashResult {
	ch := make(chan stashResult, 1)
	go func() {
		v, err := s.Stash(ctx, "mod", "ver")
		ch <- stashResult{v, err}
	}()
	return ch
}

// TestPoolCancelWhileEnqueuing ensures a caller whose context is done
// while every worker is busy returns promptly instead of blocking on the queue.
func TestPoolCancelWhileEnqueuing(t *testing.T) {
	b := &blockingStasher{started: make(chan struct{}, 2), release: make(chan struct{})}
	s := WithPool(1)(b)

	first := stashAsync(t.Context(), s)
	<-b.started // the only worker is now occupied.

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	second := stashAsync(ctx, s)

	select {
	case res := <-second:
		if !errors.Is(res.err, context.DeadlineExceeded) {
			t.Fatalf("expected context deadline error, got %v", res.err)
		}
	case <-time.After(2 * time.Second):
		close(b.release)
		t.Fatal("Stash did not return after context was done while enqueuing")
	}

	// Releasing the occupied job must neither panic, race nor deadlock.
	close(b.release)
	select {
	case res := <-first:
		if res.err != nil || res.ver != "newver" {
			t.Fatalf("unexpected first result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("first Stash did not return after release")
	}
}

// TestPoolCancelWhileWaiting ensures a caller whose job is already running
// returns promptly on context cancellation, and that the job finishing
// later does not panic or race.
func TestPoolCancelWhileWaiting(t *testing.T) {
	b := &blockingStasher{started: make(chan struct{}, 2), release: make(chan struct{})}
	s := WithPool(1)(b)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	res := stashAsync(ctx, s)
	<-b.started
	cancel()

	select {
	case r := <-res:
		if !errors.Is(r.err, context.Canceled) {
			t.Fatalf("expected context canceled error, got %v", r.err)
		}
	case <-time.After(2 * time.Second):
		close(b.release)
		t.Fatal("Stash did not return after context was cancelled while waiting")
	}

	// The worker must be free again once the abandoned job finishes.
	close(b.release)
	again := stashAsync(t.Context(), s)
	select {
	case r := <-again:
		if r.err != nil || r.ver != "newver" {
			t.Fatalf("unexpected result: %+v", r)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("pool worker was not released after abandoned job finished")
	}
}
