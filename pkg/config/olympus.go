package config

type OlympusConfig struct {
	Port              string `validate:"required"`
	StorageType       string `validate:"required"`
	WorkerType        string `validate:"required"`
	RedisQueueAddress string `validate:"required"`
}
