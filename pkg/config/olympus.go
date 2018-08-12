package config

// OlympusConfig specifies properties required by the Olympus registry
type OlympusConfig struct {
	Port              string `validate:"required" envconfig:"PORT" default:":3001"`
	StorageType       string `validate:"required" envconfig:"ATHENS_STORAGE_TYPE" default:"memory"`
	WorkerType        string `validate:"required" envconfig:"OLYMPUS_BACKGROUND_WORKER_TYPE" default:"redis"`
	RedisQueueAddress string `validate:"required" envconfig:"OLYMPUS_REDIS_QUEUE_PORT" default:":6379"`
}
