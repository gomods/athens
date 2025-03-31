package stash

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
)

// WithRedisSentinelLock returns a distributed singleflight
// with a redis cluster that utilizes sentinel for quorum and failover.
func WithRedisSentinelLock(l RedisLogger, endpoints []string, master, sentinelPassword, redisUsername, redisPassword string, checker storage.Checker, lockConfig *config.RedisLockConfig) (Wrapper, error) {
	redis.SetLogger(l)

	const op errors.Op = "stash.WithRedisSentinelLock"
	// The redis client constructor does not return an error when no endpoints
	// are provided, so we check for ourselves.
	if len(endpoints) == 0 {
		return nil, errors.E(op, "no endpoints specified")
	}
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:       master,
		SentinelAddrs:    endpoints,
		SentinelPassword: sentinelPassword,
		Username:         redisUsername,
		Password:         redisPassword,
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, errors.E(op, err)
	}

	lockOptions, err := lockOptionsFromConfig(lockConfig)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return func(s Stasher) Stasher {
		return &redisLock{client, s, checker, lockOptions}
	}, nil
}
