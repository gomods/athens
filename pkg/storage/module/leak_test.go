package module

import (
	"bytes"
	"context"
	"io"
	"runtime"
	"testing"
	"time"
)

const (
	leakIterations = 100
	leakTolerance  = 10
)

// settledGoroutines polls until the goroutine count stops changing and
// returns it. It is used to take a baseline that is not skewed by goroutines
// started by previously executed tests that are still winding down.
func settledGoroutines() int {
	last := runtime.NumGoroutine()
	stable := 0
	deadline := time.Now().Add(5 * time.Second)
	for stable < 5 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
		n := runtime.NumGoroutine()
		if n == last {
			stable++
		} else {
			stable = 0
			last = n
		}
	}
	return last
}

// assertGoroutinesReturnToBaseline polls until the goroutine count is back
// within leakTolerance of baseline, failing the test after 5 seconds.
func assertGoroutinesReturnToBaseline(t *testing.T, baseline int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var n int
	for {
		n = runtime.NumGoroutine()
		if n <= baseline+leakTolerance {
			return
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	buf := make([]byte, 1<<16)
	buf = buf[:runtime.Stack(buf, true)]
	t.Fatalf("goroutine leak: baseline %d, now %d (tolerance %d)\n%s", baseline, n, leakTolerance, buf)
}

func blockUntilDoneUploader(ctx context.Context, _, _ string, _ io.Reader) error {
	<-ctx.Done()
	return ctx.Err()
}

func blockUntilDoneDeleter(ctx context.Context, _ string) error {
	<-ctx.Done()
	return ctx.Err()
}

func TestUploadDoesNotLeakGoroutinesOnTimeout(t *testing.T) {
	baseline := settledGoroutines()
	rd := bytes.NewReader([]byte("123"))
	for range leakIterations {
		if err := Upload(context.Background(), "mx", "1.1.1", rd, rd, rd, blockUntilDoneUploader, time.Millisecond); err == nil {
			t.Fatal("expected timeout error")
		}
	}
	assertGoroutinesReturnToBaseline(t, baseline)
}

func TestUploadDoesNotLeakGoroutinesOnParentCancel(t *testing.T) {
	baseline := settledGoroutines()
	rd := bytes.NewReader([]byte("123"))
	for range leakIterations {
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(time.Millisecond, cancel)
		if err := Upload(ctx, "mx", "1.1.1", rd, rd, rd, blockUntilDoneUploader, time.Minute); err == nil {
			t.Fatal("expected cancellation error")
		}
		cancel()
	}
	assertGoroutinesReturnToBaseline(t, baseline)
}

func TestDeleteDoesNotLeakGoroutinesOnTimeout(t *testing.T) {
	baseline := settledGoroutines()
	for range leakIterations {
		if err := Delete(context.Background(), "mx", "1.1.1", blockUntilDoneDeleter, time.Millisecond); err == nil {
			t.Fatal("expected timeout error")
		}
	}
	assertGoroutinesReturnToBaseline(t, baseline)
}

func TestDeleteDoesNotLeakGoroutinesOnParentCancel(t *testing.T) {
	baseline := settledGoroutines()
	for range leakIterations {
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(time.Millisecond, cancel)
		if err := Delete(ctx, "mx", "1.1.1", blockUntilDoneDeleter, time.Minute); err == nil {
			t.Fatal("expected cancellation error")
		}
		cancel()
	}
	assertGoroutinesReturnToBaseline(t, baseline)
}
