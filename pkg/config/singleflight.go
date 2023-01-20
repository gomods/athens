package config

// SingleFlight holds the various
// backend configurations for a distributed
// lock or single flight mechanism.
type SingleFlight struct {
	Etcd          *Etcd
	Redis         *Redis
	RedisSentinel *RedisSentinel
}

// Etcd holds client side configuration
// that helps Athens connect to the
// Etcd backends.
type Etcd struct {
	Endpoints string `envconfig:"ATHENS_ETCD_ENDPOINTS"`
}

// Redis holds the client side configuration
// to connect to redis as a SingleFlight implementation.
type Redis struct {
	Endpoint   string `envconfig:"ATHENS_REDIS_ENDPOINT"`
	Password   string `envconfig:"ATHENS_REDIS_PASSWORD"`
	LockConfig *RedisLockConfig
}

// RedisSentinel is the configuration for using redis with sentinel
// for SingleFlight
type RedisSentinel struct {
	Endpoints        []string `envconfig:"ATHENS_REDIS_SENTINEL_ENDPOINTS"`
	MasterName       string   `envconfig:"ATHENS_REDIS_SENTINEL_MASTER_NAME"`
	SentinelPassword string   `envconfig:"ATHENS_REDIS_SENTINEL_PASSWORD"`
	LockConfig       *RedisLockConfig
}

type RedisLockConfig struct {
	Timeout    int `envconfig:"ATHENS_REDIS_LOCK_TIMEOUT"`
	TTL        int `envconfig:"ATHENS_REDIS_LOCK_TTL"`
	MaxRetries int `envconfig:"ATHENS_REDIS_LOCK_MAX_RETRIES"`
}

func DefaultRedisLockConfig() *RedisLockConfig {
	return &RedisLockConfig{
		TTL:        900,
		Timeout:    15,
		MaxRetries: 10,
	}
}
