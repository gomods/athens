package config

import (
	"runtime"

	"github.com/spf13/viper"
	validator "gopkg.in/go-playground/validator.v9"
)

// Config provides configuration values for all components
type Config struct {
	EnableCSRFProtection bool
	GoEnv                string         `validate:"required"`
	GoBinary             string         `validate:"required"`
	LogLevel             string         `validate:"required"`
	MaxConcurrency       int            `validate:"required"`
	MaxWorkerFails       uint           `validate:"required"`
	CloudRuntime         string         `validate:"required"`
	FilterFile           string         `validate:"required"`
	TimeoutSeconds       int            `validate:"required"`
	Proxy                *ProxyConfig   `validate:""`
	Olympus              *OlympusConfig `validate:""`
	Storage              *StorageConfig `validate:""`
}

// ParseConfig parses the given file into an athens config struct
func ParseConfig(configFile string) (*Config, error) {

	viper.SetConfigFile(configFile)

	// attempt to parse the given config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// set default values and environment variable bindings
	setDefaultsAndEnvBindings()
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	// validate all required fields have been populated
	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, err
	}
	return &config, nil
}

func setDefaultsAndEnvBindings() {
	viper.BindEnv("GoBinary", "GO_BINARY_PATH")
	viper.SetDefault("GoBinary", "go")

	viper.BindEnv("GoEnv", "GO_ENV")
	viper.SetDefault("GoEnv", "development")

	viper.BindEnv("LogLevel", "ATHENS_LOG_LEVEL")
	viper.SetDefault("LogLevel", "debug")

	viper.BindEnv("CloudRuntime", "ATHENS_CLOUD_RUNTIME")
	viper.SetDefault("CloudRuntime", "none")

	viper.BindEnv("MaxConcurrency", "ATHENS_MAX_CONCURRENCY")
	viper.SetDefault("MaxConcurrency", runtime.NumCPU())

	viper.BindEnv("MaxWorkerFails", "ATHENS_WORKER_MAX_FAILS")
	viper.SetDefault("MaxWorkerFails", 5)

	viper.BindEnv("FilterFile", "ATHENS_FILTER_FILENAME")
	viper.SetDefault("FilterFile", "filter.conf")

	viper.BindEnv("TimeoutSeconds", "ATHENS_TIMEOUT")
	viper.SetDefault("TimeoutSeconds", 300)

	viper.BindEnv("EnableCSRFProtection", "ATHENS_ENABLE_CSRF_PROTECTION")
	viper.SetDefault("EnableCSRFProtection", false)

	// Proxy Defaults
	viper.BindEnv("Proxy.StorageType", "ATHENS_STORAGE_TYPE")
	viper.SetDefault("Proxy.StorageType", "mongo")

	viper.BindEnv("Proxy.Port", "PORT")
	viper.SetDefault("Proxy.Port", ":3000")

	viper.BindEnv("Proxy.OlympusGlobalEndpoint", "OLYMPUS_GLOBAL_ENDPOINT")
	viper.SetDefault("Proxy.OlympusGlobalEndpoint", "olympus.gomods.io")

	viper.BindEnv("Proxy.RedisQueueAddress", "ATHENS_REDIS_QUEUE_PORT")
	viper.SetDefault("Proxy.RedisQueueAddress", ":6379")

	// Olympus Defaults
	viper.BindEnv("Olympus.Port", "PORT")
	viper.SetDefault("Olympus.Port", ":3001")

	viper.BindEnv("Olympus.StorageType", "ATHENS_STORAGE_TYPE")
	viper.SetDefault("Olympus.StorageType", "memory")

	viper.BindEnv("Olympus.WorkerType", "OLYMPUS_BACKGROUND_WORKER_TYPE")
	viper.SetDefault("Olympus.WorkerType", "redis")

	viper.BindEnv("Olympus.RedisQueueAddress", "OLYMPUS_REDIS_QUEUE_PORT")
	viper.SetDefault("Olympus.RedisQueueAddress", ":6379")

	// Storage defaults
	setStorageDefaults(viper.GetViper())
}
