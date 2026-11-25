package stash

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type blockingStasher struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
	calls   atomic.Int32
	fn      func(ctx context.Context, mod, ver string) (string, error)
}

func (b *blockingStasher) Stash(ctx context.Context, mod, ver string) (string, error) {
	b.calls.Add(1)
	if b.started != nil {
		b.once.Do(func() { close(b.started) })
	}
	if b.release != nil {
		select {
		case <-b.release:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	if b.fn != nil {
		return b.fn(ctx, mod, ver)
	}
	return ver, nil
}

func TestSingleFlightWaiterCancel(t *testing.T) {
	base := &blockingStasher{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	sf := WithSingleflight(base)

	ctx1, cancel1 := context.WithCancel(t.Context())
	defer cancel1()
	ctx2 := t.Context()

	err1Ch := make(chan error, 1)
	go func() {
		_, err := sf.Stash(ctx1, "example.com/mod", "v1.0.0")
		err1Ch <- err
	}()

	<-base.started

	res2Ch := make(chan struct {
		ver string
		err error
	}, 1)
	go func() {
		ver, err := sf.Stash(ctx2, "example.com/mod", "v1.0.0")
		res2Ch <- struct {
			ver string
			err error
		}{ver: ver, err: err}
	}()

	// Allow second caller to subscribe before canceling first waiter.
	time.Sleep(20 * time.Millisecond)
	cancel1()

	err1 := <-err1Ch
	require.ErrorIs(t, err1, context.Canceled)

	close(base.release)
	res2 := <-res2Ch
	require.NoError(t, res2.err)
	require.Equal(t, "v1.0.0", res2.ver)
	require.Equal(t, int32(1), base.calls.Load())
}

func TestSingleFlightKeepsDeadline(t *testing.T) {
	base := &blockingStasher{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	sf := WithSingleflight(base)

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer cancel()

	_, err := sf.Stash(ctx, "example.com/mod", "v1.0.0")
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestLocalSingleflightCoalescesBeforeDistributed(t *testing.T) {
	var distCalls atomic.Int32
	base := &blockingStasher{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	distributed := func(next Stasher) Stasher {
		return stasherFunc(func(ctx context.Context, mod, ver string) (string, error) {
			distCalls.Add(1)
			return next.Stash(ctx, mod, ver)
		})
	}

	sf := WithLocalSingleflight(distributed)(base)

	const callers = 4
	errs := make(chan error, callers)
	for range callers {
		go func() {
			_, err := sf.Stash(t.Context(), "example.com/mod", "v1.0.0")
			errs <- err
		}()
	}

	<-base.started
	time.Sleep(20 * time.Millisecond)
	close(base.release)

	for range callers {
		require.NoError(t, <-errs)
	}
	require.Equal(t, int32(1), distCalls.Load())
	require.Equal(t, int32(1), base.calls.Load())
}

func TestLocalSingleflightSharesFailureThenRetries(t *testing.T) {
	boom := errors.New("upstream boom")
	var attempt atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})

	base := stasherFunc(func(_ context.Context, _, ver string) (string, error) {
		if attempt.Add(1) == 1 {
			close(started)
			<-release
			return "", boom
		}
		return ver, nil
	})
	sf := WithLocalSingleflight(func(s Stasher) Stasher { return s })(base)

	err1Ch := make(chan error, 1)
	err2Ch := make(chan error, 1)
	go func() {
		_, err := sf.Stash(t.Context(), "example.com/mod", "v1.0.0")
		err1Ch <- err
	}()
	<-started
	go func() {
		_, err := sf.Stash(t.Context(), "example.com/mod", "v1.0.0")
		err2Ch <- err
	}()
	time.Sleep(20 * time.Millisecond)
	close(release)

	require.ErrorIs(t, <-err1Ch, boom)
	require.ErrorIs(t, <-err2Ch, boom)

	ver, err := sf.Stash(t.Context(), "example.com/mod", "v1.0.0")
	require.NoError(t, err)
	require.Equal(t, "v1.0.0", ver)
	require.Equal(t, int32(2), attempt.Load())
}

type stasherFunc func(ctx context.Context, mod, ver string) (string, error)

func (f stasherFunc) Stash(ctx context.Context, mod, ver string) (string, error) {
	return f(ctx, mod, ver)
}
