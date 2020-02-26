package stash

import (
	"github.com/go-redis/redis/v7"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
)

// WithRedisSentinelLock returns a distributed singleflight
// with a redis cluster that utilizes sentinel for quorum and failover
func WithRedisSentinelLock(endpoints []string, master, password string, checker storage.Checker) (Wrapper, error) {
	const op errors.Op = "stash.WithRedisSentinelLock"
	// The redis client constructor does not return an error when no endpoints
	// are provided, so we check for ourselves.
	if len(endpoints) == 0 {
		return nil, errors.E(op, "no endpoints specified")
	}
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:       master,
		SentinelAddrs:    endpoints,
		SentinelPassword: password,
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, errors.E(op, err)
	}
	return func(s Stasher) Stasher {
		return &redisLock{client, s, checker}
	}, nil
}
