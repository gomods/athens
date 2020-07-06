package stash

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_lockerLock_Stash(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		errs := make(chan error, 1)
		ll := &lockerLock{
			locker: mockLocker(func(ctx context.Context, name string) (releaseErrs <-chan error, err error) {
				assert.Equal(t, "mod@ver", name)
				go func() {
					<- ctx.Done()
					close(errs)
				}()
				return errs, nil
			}),
			checker: mockChecker(func(ctx context.Context, module, version string) (bool, error) {
				assert.Equal(t, "mod", module)
				assert.Equal(t, "ver", version)
				return true, nil
			}),
			stasher: mockStasher(func(ctx context.Context, mod, ver string) (string, error) {
				assert.Fail(t, "this should not be called")
				return "", nil
			}),
		}
		ctx := context.Background()
		got, err := ll.Stash(ctx, "mod", "ver")
		require.NoError(t, err)
		require.Equal(t, "ver", got)
	})

	t.Run("doesn't exist", func(t *testing.T) {
		errs := make(chan error, 1)
		ll := &lockerLock{
			locker: mockLocker(func(ctx context.Context, name string) (releaseErrs <-chan error, err error) {
				assert.Equal(t, "mod@ver", name)
				go func() {
					<- ctx.Done()
					close(errs)
				}()
				return errs, nil
			}),
			checker: mockChecker(func(ctx context.Context, module, version string) (bool, error) {
				assert.Equal(t, "mod", module)
				assert.Equal(t, "ver", version)
				return false, nil
			}),
			stasher: mockStasher(func(ctx context.Context, mod, ver string) (string, error) {
				assert.Equal(t, "mod", mod)
				assert.Equal(t, "ver", ver)
				return "newVer", nil
			}),
		}
		ctx := context.Background()
		got, err := ll.Stash(ctx, "mod", "ver")
		require.NoError(t, err)
		require.Equal(t, "newVer", got)
	})

	t.Run("can't get lock", func(t *testing.T) {
		ll := &lockerLock{
			locker: mockLocker(func(ctx context.Context, name string) (releaseErrs <-chan error, err error) {
				assert.Equal(t, "mod@ver", name)
				return nil, fmt.Errorf("error")
			}),
			checker: mockChecker(func(ctx context.Context, module, version string) (bool, error) {
				assert.Fail(t, "this should not be called")
				return false, nil
			}),
			stasher: mockStasher(func(ctx context.Context, mod, ver string) (string, error) {
				assert.Fail(t, "this should not be called")
				return "newVer", nil
			}),
		}
		ctx := context.Background()
		_, err := ll.Stash(ctx, "mod", "ver")
		require.EqualError(t, err, "error")
	})

	t.Run("checker error", func(t *testing.T) {
		errs := make(chan error, 1)
		ll := &lockerLock{
			locker: mockLocker(func(ctx context.Context, name string) (releaseErrs <-chan error, err error) {
				assert.Equal(t, "mod@ver", name)
				return errs, nil
			}),
			checker: mockChecker(func(ctx context.Context, module, version string) (bool, error) {
				return false, fmt.Errorf("foo")
			}),
			stasher: mockStasher(func(ctx context.Context, mod, ver string) (string, error) {
				assert.Fail(t, "this should not be called")
				return "", nil
			}),
		}
		ctx := context.Background()
		_, err := ll.Stash(ctx, "mod", "ver")
		require.EqualError(t, err, "foo")
	})

	t.Run("context closes during check", func(t *testing.T) {
		errs := make(chan error, 1)
		ctx, cancel := context.WithCancel(context.Background())
		ll := &lockerLock{
			locker: mockLocker(func(ctx context.Context, name string) (releaseErrs <-chan error, err error) {
				assert.Equal(t, "mod@ver", name)
				return errs, nil
			}),
			checker: mockChecker(func(ctx context.Context, module, version string) (bool, error) {
				assert.Equal(t, "mod", module)
				assert.Equal(t, "ver", version)
				cancel()
				go func() {
					time.Sleep(10 * time.Millisecond)
					close(errs)
				}()
				return false, nil
			}),
			stasher: mockStasher(func(ctx context.Context, mod, ver string) (string, error) {
				assert.Fail(t, "this should not be called")
				return "", nil
			}),
		}
		_, err := ll.Stash(ctx, "mod", "ver")
		require.EqualError(t, err, context.Canceled.Error())
	})

	t.Run("loses lock during check", func(t *testing.T) {
		errs := make(chan error, 1)
		ll := &lockerLock{
			locker: mockLocker(func(ctx context.Context, name string) (releaseErrs <-chan error, err error) {
				assert.Equal(t, "mod@ver", name)
				return errs, nil
			}),
			checker: mockChecker(func(ctx context.Context, module, version string) (bool, error) {
				assert.Equal(t, "mod", module)
				assert.Equal(t, "ver", version)
				errs <- ErrUnexpectedRelease
				close(errs)
				time.Sleep(10 * time.Millisecond)
				return false, nil
			}),
			stasher: mockStasher(func(ctx context.Context, mod, ver string) (string, error) {
				assert.Fail(t, "this should not be called")
				return "", nil
			}),
		}
		ctx := context.Background()
		_, err := ll.Stash(ctx, "mod", "ver")
		require.EqualError(t, err, ErrUnexpectedRelease.Error())
	})

	t.Run("stash error", func(t *testing.T) {
		errs := make(chan error, 1)
		ll := &lockerLock{
			locker: mockLocker(func(ctx context.Context, name string) (releaseErrs <-chan error, err error) {
				assert.Equal(t, "mod@ver", name)
				go func() {
					<- ctx.Done()
					close(errs)
				}()
				return errs, nil
			}),
			checker: mockChecker(func(ctx context.Context, module, version string) (bool, error) {
				assert.Equal(t, "mod", module)
				assert.Equal(t, "ver", version)
				return false, nil
			}),
			stasher: mockStasher(func(ctx context.Context, mod, ver string) (string, error) {
				assert.Equal(t, "mod", mod)
				assert.Equal(t, "ver", ver)
				return "newVer", fmt.Errorf("foo")
			}),
		}
		ctx := context.Background()
		got, err := ll.Stash(ctx, "mod", "ver")
		require.EqualError(t, err, "foo")
		require.Equal(t, "ver", got)
	})
}

func TestWithLocker(t *testing.T) {
	var lockChans []chan error
	var lockChansMux sync.Mutex
	var lwg sync.WaitGroup
	lwg.Add(5)
	go func() {
		lwg.Wait()
		time.Sleep(100 * time.Millisecond)
		lockChansMux.Lock()
		for _, ch := range lockChans {
			close(ch)
		}
		lockChansMux.Unlock()
	}()
	var lockOnce sync.Once
	lckr := mockLocker(func(_ context.Context, name string) (releaseErrs <-chan error, err error) {
		firstTime := false
		lockOnce.Do(func() {
			firstTime = true
		})
		ch := make(chan error, 1)
		lockChansMux.Lock()
		lockChans = append(lockChans, ch)
		lockChansMux.Unlock()
		if firstTime {
			lwg.Done()
			return ch, nil
		}
		lwg.Done()
		lwg.Wait()
		return ch, nil
	})
	var checkerCount int
	var checkerMux sync.Mutex
	checker := mockChecker(func(_ context.Context, mod, ver string) (bool, error) {
		t.Helper()
		checkerMux.Lock()
		defer checkerMux.Unlock()
		assert.Equal(t, "mod", mod)
		assert.Equal(t, "ver", ver)
		checkerCount++
		if checkerCount > 1 {
			return true, nil
		}
		return false, nil
	})
	stashCount := 0
	ms := mockStasher(func(_ context.Context, mod, ver string) (string, error) {
		t.Helper()
		assert.Equal(t, "mod", mod)
		assert.Equal(t, "ver", ver)
		stashCount++
		time.Sleep(100 * time.Millisecond)
		return "newVer", nil
	})
	wrapper := withLocker(lckr, checker)
	s := wrapper(ms)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			_, err := s.Stash(context.Background(), "mod", "ver")
			require.NoError(t, err)
			wg.Done()
		}()
	}
	wg.Wait()
	require.Equal(t, 1, stashCount)
	require.Equal(t, 5, checkerCount)
}

func Test_lockHolder_holdAndRelease(t *testing.T) {
	t.Run("refreshes until ctx.Done()", func(t *testing.T) {
		refreshCount := 0
		releaseCount := 0
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lh := &lockHolder{
			pingInterval: 1 * time.Millisecond,
			refresh: func(refreshCtx context.Context) error {
				assert.NoError(t, refreshCtx.Err())
				refreshCount++
				if refreshCount == 5 {
					cancel()
				}
				return nil
			},
			release: func() error {
				releaseCount++
				require.Equal(t, context.Canceled, ctx.Err())
				return nil
			},
		}
		errs := make(chan error, 1)
		lh.holdAndRelease(ctx, errs)
		require.NoError(t, <-errs)
		require.Equal(t, 5, refreshCount)
		require.Equal(t, 1, releaseCount)
	})

	t.Run("refresh error", func(t *testing.T) {
		refreshCount := 0
		releaseCount := 0
		ctx := context.Background()
		lh := &lockHolder{
			pingInterval: 1 * time.Millisecond,
			refresh: func(refreshCtx context.Context) error {
				refreshCount++
				if refreshCount == 5 {
					return assert.AnError
				}
				return nil
			},
			release: func() error {
				releaseCount++
				require.NoError(t, ctx.Err())
				return nil
			},
		}
		errs := make(chan error, 1)
		lh.holdAndRelease(ctx, errs)
		err := <-errs
		require.EqualError(t, err, assert.AnError.Error())
		require.Equal(t, 5, refreshCount)
		require.Equal(t, 1, releaseCount)
	})

	t.Run("release error", func(t *testing.T) {
		refreshCount := 0
		releaseCount := 0
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var mintime time.Time
		lh := &lockHolder{
			pingInterval: 1 * time.Millisecond,
			ttl:          10 * time.Millisecond,
			refresh: func(refreshCtx context.Context) error {
				refreshCount++
				if refreshCount == 5 {
					cancel()
				}
				return nil
			},
			release: func() error {
				releaseCount++
				require.Equal(t, context.Canceled, ctx.Err())
				mintime = time.Now().Add(10 * time.Millisecond)
				return assert.AnError
			},
		}
		errs := make(chan error, 1)
		lh.holdAndRelease(ctx, errs)
		require.NoError(t, <-errs)
		require.Equal(t, 5, refreshCount)
		require.Equal(t, 1, releaseCount)
		require.True(t, time.Now().After(mintime))
	})
}

type mockLocker func(ctx context.Context, name string) (releaseErrs <-chan error, err error)

func (m mockLocker) lock(ctx context.Context, name string) (releaseErrs <-chan error, err error) {
	return m(ctx, name)
}
