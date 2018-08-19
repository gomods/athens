package config

// ProxyConfig specifies the properties required to run the proxy
type ProxyConfig struct {
	StorageType           string `validate:"required" envconfig:"ATHENS_STORAGE_TYPE"`
	OlympusGlobalEndpoint string `validate:"required" envconfig:"OLYMPUS_GLOBAL_ENDPOINT"`
	Port                  string `validate:"required" envconfig:"PORT"`
	RedisQueueAddress     string `validate:"required" envconfig:"ATHENS_REDIS_QUEUE_PORT"`
	FilterOff             *bool  `validate:"required" envconfig:"PROXY_FILTER_OFF"`
}

func setProxyDefaults(p *ProxyConfig) *ProxyConfig {
	if p == nil {
		p = &ProxyConfig{}
	}
	overrideDefaultStr(&p.StorageType, "memory")
	overrideDefaultStr(&p.OlympusGlobalEndpoint, "http://localhost:3001")
	overrideDefaultStr(&p.Port, ":3000")
	overrideDefaultStr(&p.RedisQueueAddress, ":6379")
	filter := true
	if p.FilterOff == nil {
		p.FilterOff = &filter
	}
	return p
}
