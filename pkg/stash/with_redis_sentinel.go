package stash

import (
	"github.com/go-redis/redis"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
)

// WithRedisSentinelLock returns a distributed singleflight
// with a redis cluster that utilizes sentinel for quorum and failover
func WithRedisSentinelLock(endpoints []string, master, password string, checker storage.Checker) (Wrapper, error) {
	const op errors.Op = "stash.WithRedisSentinelLock"
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    master,
		SentinelAddrs: endpoints,
		Password:      password,
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, errors.E(op, err)
	}
	return func(s Stasher) Stasher {
		return &redisLock{client, s, checker}
	}, nil
}
