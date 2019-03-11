package lock

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

var luaRefresh = redis.NewScript(`if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("pexpire", KEYS[1], ARGV[2]) else return 0 end`)
var luaRelease = redis.NewScript(`if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`)

var emptyCtx = context.Background()

// ErrLockNotObtained may be returned by Obtain() and Run()
// if a lock could not be obtained.
var (
	ErrLockUnlockFailed     = errors.New("lock unlock failed")
	ErrLockNotObtained      = errors.New("lock not obtained")
	ErrLockDurationExceeded = errors.New("lock duration exceeded")
)

// RedisClient is a minimal client interface.
type RedisClient interface {
	SetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Eval(script string, keys []string, args ...interface{}) *redis.Cmd
	EvalSha(sha1 string, keys []string, args ...interface{}) *redis.Cmd
	ScriptExists(scripts ...string) *redis.BoolSliceCmd
	ScriptLoad(script string) *redis.StringCmd
}

// Locker allows (repeated) distributed locking.
type Locker struct {
	client RedisClient
	key    string
	opts   Options

	token string
	mutex sync.Mutex
}

// Run runs a callback handler with a Redis lock. It may return ErrLockNotObtained
// if a lock was not successfully acquired.
func Run(client RedisClient, key string, opts *Options, handler func()) error {
	locker, err := Obtain(client, key, opts)
	if err != nil {
		return err
	}

	sem := make(chan struct{})
	go func() {
		handler()
		close(sem)
	}()

	select {
	case <-sem:
		return locker.Unlock()
	case <-time.After(locker.opts.LockTimeout):
		return ErrLockDurationExceeded
	}
}

// Obtain is a shortcut for New().Lock(). It may return ErrLockNotObtained
// if a lock was not successfully acquired.
func Obtain(client RedisClient, key string, opts *Options) (*Locker, error) {
	locker := New(client, key, opts)
	if ok, err := locker.Lock(); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrLockNotObtained
	}
	return locker, nil
}

// New creates a new distributed locker on a given key.
func New(client RedisClient, key string, opts *Options) *Locker {
	var o Options
	if opts != nil {
		o = *opts
	}
	o.normalize()

	return &Locker{client: client, key: key, opts: o}
}

// IsLocked returns true if a lock is still being held.
func (l *Locker) IsLocked() bool {
	l.mutex.Lock()
	locked := l.token != ""
	l.mutex.Unlock()

	return locked
}

// Lock applies the lock, don't forget to defer the Unlock() function to release the lock after usage.
func (l *Locker) Lock() (bool, error) {
	return l.LockWithContext(emptyCtx)
}

// LockWithContext is like Lock but allows to pass an additional context which allows cancelling
// lock attempts prematurely.
func (l *Locker) LockWithContext(ctx context.Context) (bool, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.token != "" {
		return l.refresh(ctx)
	}
	return l.create(ctx)
}

// Unlock releases the lock
func (l *Locker) Unlock() error {
	l.mutex.Lock()
	err := l.release()
	l.mutex.Unlock()

	return err
}

// Helpers

func (l *Locker) create(ctx context.Context) (bool, error) {
	l.reset()

	// Create a random token
	token, err := randomToken()
	if err != nil {
		return false, err
	}

	// Calculate the timestamp we are willing to wait for
	attempts := l.opts.RetryCount + 1
	var retryDelay *time.Timer

	for {

		// Try to obtain a lock
		ok, err := l.obtain(token)
		if err != nil {
			return false, err
		} else if ok {
			l.token = token
			return true, nil
		}

		if attempts--; attempts <= 0 {
			return false, nil
		}

		if retryDelay == nil {
			retryDelay = time.NewTimer(l.opts.RetryDelay)
			defer retryDelay.Stop()
		} else {
			retryDelay.Reset(l.opts.RetryDelay)
		}

		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-retryDelay.C:
		}
	}
}

func (l *Locker) refresh(ctx context.Context) (bool, error) {
	ttl := strconv.FormatInt(int64(l.opts.LockTimeout/time.Millisecond), 10)
	status, err := luaRefresh.Run(l.client, []string{l.key}, l.token, ttl).Result()
	if err != nil {
		return false, err
	} else if status == int64(1) {
		return true, nil
	}
	return l.create(ctx)
}

func (l *Locker) obtain(token string) (bool, error) {
	ok, err := l.client.SetNX(l.key, token, l.opts.LockTimeout).Result()
	if err == redis.Nil {
		err = nil
	}
	return ok, err
}

func (l *Locker) release() error {
	defer l.reset()

	res, err := luaRelease.Run(l.client, []string{l.key}, l.token).Result()
	if err == redis.Nil {
		return ErrLockUnlockFailed
	}

	if i, ok := res.(int64); !ok || i != 1 {
		return ErrLockUnlockFailed
	}

	return err
}

func (l *Locker) reset() {
	l.token = ""
}

func randomToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}
