package config

// SingleFlight holds the various backend configurations for a distributed lock or single flight mechanism.
type SingleFlight struct {
	Etcd          *EtcdSingleFlight
	Redis         *RedisSingleFlight
	RedisSentinel *RedisSentinelSingleFlight
	GCP           *GCPSingleFlight
}

// EtcdSingleFlight holds client side configuration
// that helps Athens connect to the Etcd backends.
type EtcdSingleFlight struct {
	Endpoints string `envconfig:"ATHENS_ETCD_ENDPOINTS"`
}

// RedisSingleFlight holds the client side configuration
// to connect to redis as a SingleFlight implementation.
type RedisSingleFlight struct {
	Endpoint   string `envconfig:"ATHENS_REDIS_ENDPOINT"`
	Password   string `envconfig:"ATHENS_REDIS_PASSWORD"`
	LockConfig *RedisLockConfig
}

// RedisSentinelSingleFlight is the configuration for using redis with sentinel
// for SingleFlight.
type RedisSentinelSingleFlight struct {
	Endpoints        []string `envconfig:"ATHENS_REDIS_SENTINEL_ENDPOINTS"`
	MasterName       string   `envconfig:"ATHENS_REDIS_SENTINEL_MASTER_NAME"`
	SentinelPassword string   `envconfig:"ATHENS_REDIS_SENTINEL_PASSWORD"`
	RedisUsername    string   `envconfig:"ATHENS_REDIS_USERNAME"`
	RedisPassword    string   `envconfig:"ATHENS_REDIS_PASSWORD"`
	LockConfig       *RedisLockConfig
}

// RedisLockConfig is the configuration for redis locking.
type RedisLockConfig struct {
	Timeout    int `envconfig:"ATHENS_REDIS_LOCK_TIMEOUT"`
	TTL        int `envconfig:"ATHENS_REDIS_LOCK_TTL"`
	MaxRetries int `envconfig:"ATHENS_REDIS_LOCK_MAX_RETRIES"`
}

// DefaultRedisLockConfig returns the default redis locking configuration.
func DefaultRedisLockConfig() *RedisLockConfig {
	return &RedisLockConfig{
		TTL:        900,
		Timeout:    15,
		MaxRetries: 10,
	}
}

// GCPSingleFlight is the configuration for GCP locking.
type GCPSingleFlight struct {
	StaleThreshold int `envconfig:"ATHENS_GCP_STALE_THRESHOLD"`
}
