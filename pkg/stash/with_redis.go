package stash

import (
	"context"
	"time"

	lock "github.com/bsm/redislock"
	"github.com/go-redis/redis/v7"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

// WithRedisLock returns a distributed singleflight
// using an redis cluster. If it cannot connect, it will return an error.
func WithRedisLock(endpoint string, password string, getter storage.Getter) (Wrapper, error) {
	const op errors.Op = "stash.WithRedisLock"
	client := redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     endpoint,
		Password: password,
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, errors.E(op, err)
	}

	return func(s Stasher) Stasher {
		return &redisLock{client, s, getter}
	}, nil
}

type redisLock struct {
	client  *redis.Client
	stasher Stasher
	getter  storage.Getter
}

func (s *redisLock) Stash(ctx context.Context, mod, ver string) (newVer string, err error) {
	const op errors.Op = "redis.Stash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	mv := config.FmtModVer(mod, ver)

	// Obtain a new lock with default settings
	lock, err := lock.Obtain(s.client, mv, time.Minute*5, &lock.Options{
		RetryStrategy: lock.LimitRetry(lock.LinearBackoff(time.Second), 60*5),
	})
	if err != nil {
		return ver, errors.E(op, err)
	}
	defer func() {
		const op errors.Op = "redis.Release"
		lockErr := lock.Release()
		if err == nil && lockErr != nil {
			err = errors.E(op, lockErr)
		}
	}()
	_, err = s.getter.Info(ctx, mod, ver)
	if err == nil {
		return ver, nil
	}
	if !errors.Is(err, errors.KindNotFound) {
		return ver, errors.E(op, err)
	}
	newVer, err = s.stasher.Stash(ctx, mod, ver)
	if err != nil {
		return ver, errors.E(op, err)
	}
	return newVer, nil
}
