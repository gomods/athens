package stash

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

const (
	defaultGetLockTimeout = 5 * time.Minute
	defaultPingInterval   = 10 * time.Second
)

type locker interface {
	// lock obtains a lock to be used for single flight.
	// The lock will be released when ctx is canceled. When the lock is released whether intentionally or not,
	// releaseErrs will receive an error. If that error is nil, the lock was successfully released as a response to ctx.
	// Returns:
	//	 releaseErrs:	A channel that emits an error when the lock is released.
	//   err: 			Any error creating the lock. If err is non-nil, no lock is created.
	lock(ctx context.Context, name string) (releaseErrs <-chan error, err error)
}

var ErrUnexpectedRelease = fmt.Errorf("lock was unexpectedly released")

func withLocker(l locker, checker storage.Checker) Wrapper {
	return func(s Stasher) Stasher {
		return &lockerLock{
			locker:  l,
			stasher: s,
			checker: checker,
		}
	}
}

type lockerLock struct {
	locker  locker
	stasher Stasher
	checker storage.Checker
}

func (s *lockerLock) Stash(ctx context.Context, mod, ver string) (string, error) {
	const op errors.Op = "lockerLock.Stash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mv := fmt.Sprintf("%s@%s", mod, ver)
	releaseErrs, err := s.locker.lock(ctx, mv)
	if err != nil {
		return ver, errors.E(op, err)
	}
	var releaseErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		releaseErr = <-releaseErrs
		cancel()
		wg.Done()
	}()
	exists, err := s.checker.Exists(ctx, mod, ver)
	if err != nil {
		return ver, errors.E(op, err)
	}
	if ctx.Err() != nil {
		wg.Wait()
		if releaseErr != nil {
			return ver,  errors.E(op, releaseErr)
		}
		return ver, ctx.Err()
	}
	var newVer string
	if exists {
		newVer = ver
	} else {
		newVer, err = s.stasher.Stash(ctx, mod, ver)
		if err != nil {
			return ver, errors.E(op, err)
		}
	}
	cancel()
	wg.Wait()
	return newVer, nil
}

type lockHolder struct {
	pingInterval time.Duration
	ttl          time.Duration
	refresh      func(refreshCtx context.Context) error
	release      func() error
}

func (l *lockHolder) holdAndRelease(ctx context.Context, errs chan error) {
	pingInterval := l.pingInterval
	if pingInterval == 0 {
		pingInterval = defaultPingInterval
	}
	ttl := l.ttl
	if ttl == 0 {
		ttl = pingInterval * 2
	}
	expiry := time.Now().Add(l.ttl)
	ticker := time.NewTicker(l.pingInterval)
	defer ticker.Stop()
	var err error
	func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err = l.refresh(ctx)
				if err != nil {
					return
				}
				expiry = time.Now().Add(l.ttl)
			}
		}
	}()
	releaseErr := l.release()
	if releaseErr != nil {
		time.Sleep(time.Until(expiry))
	}
	if err != nil {
		errs <- err
	}
	close(errs)
}
