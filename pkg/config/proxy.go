package config

type ProxyConfig struct {
	StorageType           string `validate:"required"`
	OlympusGlobalEndpoint string `validate:"required"`
	Port                  string `validate:"required"`
	RedisQueueAddress     string `validate:"required"`
}
