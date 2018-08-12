package config

// ProxyConfig specifies the properties required to run the proxy
type ProxyConfig struct {
	StorageType           string `validate:"required" envconfig:"ATHENS_STORAGE_TYPE" default:"mongo"`
	OlympusGlobalEndpoint string `validate:"required" envconfig:"OLYMPUS_GLOBAL_ENDPOINT" default:"olympus.gomods.io"`
	Port                  string `validate:"required" envconfig:"PORT" default:":3000"`
	RedisQueueAddress     string `validate:"required" envconfig:"ATHENS_REDIS_QUEUE_PORT" default:":6379"`
}
