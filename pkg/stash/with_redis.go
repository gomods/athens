package stash

import (
	"context"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

const (
	defaultRedisTTL        = 300 * time.Second
	defaultRedisTimeout    = 300 * time.Second
	defaultRedisMaxRetries = 300
)

// WithRedisLock returns a distributed singleflight
// using a redis cluster. If it cannot connect, it will return an error.
func WithRedisLock(endpoint string, password string, checker storage.Checker, lockConfig *config.RedisLockConfig) (Wrapper, error) {
	const op errors.Op = "stash.WithRedisLock"
	client := redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     endpoint,
		Password: password,
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, errors.E(op, err)
	}

	return func(s Stasher) Stasher {
		return &redisLock{client, s, checker, lockOptionsFromConfig(lockConfig)}
	}, nil
}

func lockOptionsFromConfig(lockConfig *config.RedisLockConfig) redisLockOptions {
	lockOptions := defaultRedisLockOptions()
	if lockConfig.TTL > 0 {
		lockOptions.ttl = time.Duration(lockConfig.TTL) * time.Second
	}
	if lockConfig.Timeout > 0 {
		lockOptions.timeout = time.Duration(lockConfig.Timeout) * time.Second
	}
	if lockConfig.MaxRetries > 0 {
		lockOptions.maxRetries = lockConfig.MaxRetries
	}
	return lockOptions
}

func defaultRedisLockOptions() redisLockOptions {
	return redisLockOptions{
		ttl:        defaultRedisTTL,
		timeout:    defaultRedisTimeout,
		maxRetries: defaultRedisMaxRetries,
	}
}

type redisLockOptions struct {
	ttl        time.Duration
	timeout    time.Duration
	maxRetries int
}

type redisLock struct {
	client  *redis.Client
	stasher Stasher
	checker storage.Checker
	options redisLockOptions
}

func (s *redisLock) Stash(ctx context.Context, mod, ver string) (newVer string, err error) {
	const op errors.Op = "redis.Stash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	mv := config.FmtModVer(mod, ver)
	lockCtx, cancel := context.WithTimeout(ctx, s.options.timeout)
	defer cancel()

	// Obtain a new lock using lock options
	lock, err := redislock.Obtain(lockCtx, s.client, mv, s.options.ttl, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(time.Second), s.options.maxRetries),
	})
	if err != nil {
		return ver, errors.E(op, err)
	}
	defer func() {
		const op errors.Op = "redis.Release"
		lockErr := lock.Release(ctx)
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
