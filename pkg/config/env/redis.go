package env

import (
	"strconv"

	"github.com/gobuffalo/envy"
)

// RedisQueuePortWithDefault returns Redis queue port used by workers defined by ATHENS_REDIS_QUEUE_PORT.
// Standard port is 6379
func RedisQueuePortWithDefault(value string) string {
	return envy.Get("ATHENS_REDIS_QUEUE_PORT", value)
}

// RedisMockInMem determines whether an in-memory worker is used to mock Redis for Proxy (Athens)
func RedisMockInMem() bool {
	boolStr := envy.Get("ATHENS_REDIS_MOCK_IN_MEM", "false")
	enable, err := strconv.ParseBool(boolStr)
	if err != nil {
		return false
	}
	return enable
}

// OlympusRedisQueuePortWithDefault returns Redis queue port used by workers defined by OLYMPUS_REDIS_QUEUE_PORT.
// Standard port is 6379
func OlympusRedisQueuePortWithDefault(value string) string {
	return envy.Get("OLYMPUS_REDIS_QUEUE_PORT", value)
}
