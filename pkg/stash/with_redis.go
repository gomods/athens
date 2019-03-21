package stash

import (
	"context"
	"time"

	lock "github.com/bsm/redis-lock"
	"github.com/go-redis/redis"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

// WithRedisLock returns a distributed singleflight
// using an redis cluster. If it cannot connect, it will return an error.
func WithRedisLock(endpoint string, checker storage.Checker) (Wrapper, error) {
	const op errors.Op = "stash.WithRedisLock"
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    endpoint,
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, errors.E(op, err)
	}

	return func(s Stasher) Stasher {
		return &redisLock{client, s, checker}
	}, nil
}

type redisLock struct {
	client  *redis.Client
	stasher Stasher
	checker storage.Checker
}

func (s *redisLock) Stash(ctx context.Context, mod, ver string) (newVer string, err error) {
	const op errors.Op = "redis.Stash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	mv := config.FmtModVer(mod, ver)

	// Obtain a new lock with default settings
	lock, err := lock.Obtain(s.client, mv, &lock.Options{
		LockTimeout: time.Minute * 5,
		RetryCount:  60 * 5,
		RetryDelay:  time.Second,
	})
	if err != nil {
		return ver, errors.E(op, err)
	}
	defer func() {
		const op errors.Op = "redis.Unlock"
		lockErr := lock.Unlock()
		if err == nil && lockErr != nil {
			err = errors.E(op, lockErr)
		}
	}()
	ok, err := s.checker.Exists(ctx, mod, ver)
	if err != nil {
		return ver, errors.E(op, err)
	}
	if ok {
		return ver, nil
	}
	newVer, err = s.stasher.Stash(ctx, mod, ver)
	if err != nil {
		return ver, errors.E(op, err)
	}
	return newVer, nil
}
