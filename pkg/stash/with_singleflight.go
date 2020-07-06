package stash

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gomods/athens/pkg/storage"
)

// WithSingleflight returns a singleflight stasher.
// This two clients make two subsequent
// requests to stash a module, then
// it will only do it once and give the first
// response to both the first and the second client.
func WithSingleflight(checker storage.Checker) Wrapper {
	lckr := &memoryLock{
		locks: map[string]bool{},
	}
	return withLocker(lckr, checker)
}

type memoryLock struct {
	mu    sync.Mutex
	locks map[string]bool
}

func (m *memoryLock) lock(ctx context.Context, name string) (releaseErrs <-chan error, err error) {
	timer := time.NewTimer(defaultGetLockTimeout)
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()
	for {
		select {
		case <-timer.C:
			return nil, fmt.Errorf("timed out waiting for lock")
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		m.mu.Lock()
		if m.locks[name] {
			m.mu.Unlock()
			time.Sleep(time.Millisecond)
			continue
		}
		m.locks[name] = true
		m.mu.Unlock()
		break
	}
	errs := make(chan error, 1)
	go func() {
		<-ctx.Done()
		m.mu.Lock()
		defer m.mu.Unlock()
		delete(m.locks, name)
		close(errs)
	}()
	return errs, nil
}
