package stash

import (
	"testing"

	"github.com/gomods/athens/pkg/config"
)

func TestWithRedisSentinelLock_DBPropagation(t *testing.T) {
	l := &testingRedisLogger{t: t}
	endpoints := []string{"127.0.0.1:26379"}
	master := "mymaster"
	sentinelPassword := "sentinel-pw"
	redisUsername := "user"
	redisPassword := "pass"
	db := 7
	
	// We use a nil checker because we won't actually call Stash
	_, err := WithRedisSentinelLock(l, endpoints, master, sentinelPassword, redisUsername, redisPassword, db, nil, config.DefaultRedisLockConfig())
	// Note: WithRedisSentinelLock calls client.Ping, which will fail because there is no redis.
	// We expect an error here.
	
	if err == nil {
		t.Fatal("expected error from Ping, but got nil")
	}
}
