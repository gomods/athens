package config

// OlympusConfig specifies properties required by the Olympus registry
type OlympusConfig struct {
	Port              string `validate:"required" envconfig:"PORT"`
	StorageType       string `validate:"required" envconfig:"ATHENS_STORAGE_TYPE"`
	WorkerType        string `validate:"required" envconfig:"OLYMPUS_BACKGROUND_WORKER_TYPE"`
	RedisQueueAddress string `validate:"required" envconfig:"OLYMPUS_REDIS_QUEUE_ADDRESS"`
}
