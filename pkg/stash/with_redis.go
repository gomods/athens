package stash

import (
	"context"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v7"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
)

// WithRedisLock returns a distributed singleflight
// using an redis cluster. If it cannot connect, it will return an error.
func WithRedisLock(endpoint string, password string, checker storage.Checker) (Wrapper, error) {
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
	lckr := &redisLock{client: client}
	return withLocker(lckr, checker), nil
}

type redisLock struct {
	client *redis.Client
}

func (l *redisLock) lock(ctx context.Context, name string) (releaseErrs <-chan error, err error) {
	ttl := defaultPingInterval * 2
	lock, err := redislock.Obtain(l.client, name, ttl, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(time.Second), 60*5),
	})
	if err != nil {
		return nil, err
	}
	holder := &lockHolder{
		pingInterval: defaultPingInterval,
		ttl:          ttl,
		refresh: func(_ context.Context) error {
			return lock.Refresh(ttl, nil)
		},
		release: lock.Release,
	}
	errs := make(chan error, 1)
	go holder.holdAndRelease(ctx, errs)
	return errs, nil
}
