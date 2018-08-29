package config

import (
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/envconfig"
	validator "gopkg.in/go-playground/validator.v9"
)

// Config provides configuration values for all components
type Config struct {
	GoEnv                string         `validate:"required" envconfig:"GO_ENV"`
	GoBinary             string         `validate:"required" envconfig:"GO_BINARY_PATH"`
	LogLevel             string         `validate:"required" envconfig:"ATHENS_LOG_LEVEL"`
	MaxConcurrency       int            `validate:"required" envconfig:"ATHENS_MAX_CONCURRENCY"`
	MaxWorkerFails       uint           `validate:"required" envconfig:"ATHENS_MAX_WORKER_FAILS"`
	CloudRuntime         string         `validate:"required" envconfig:"ATHENS_CLOUD_RUNTIME"`
	FilterFile           string         `validate:"required" envconfig:"ATHENS_FILTER_FILE"`
	Timeout              int            `validate:"required"`
	EnableCSRFProtection bool           `envconfig:"ATHENS_ENABLE_CSRF_PROTECTION"`
	Proxy                *ProxyConfig   `validate:""`
	Olympus              *OlympusConfig `validate:""`
	Storage              *StorageConfig `validate:""`
}

// ParseConfigFile parses the given file into an athens config struct
func ParseConfigFile(configFile string) (*Config, error) {

	var config Config
	// attempt to read the given config file
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return nil, err
	}

	// override values with environment variables if specified
	if err := envOverride(&config); err != nil {
		return nil, err
	}

	// set default values
	setRuntimeDefaults(&config)

	// If not defined, set storage timeouts to global timeout
	setStorageTimeouts(config.Storage, config.Timeout)

	// delete invalid storage backend configs
	// envconfig initializes *all* struct pointers, even if there are no corresponding defaults or env variables
	// this method prunes all such invalid configurations
	deleteInvalidStorageConfigs(config.Storage)

	// validate all required fields have been populated
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	return &config, nil
}

func setRuntimeDefaults(config *Config) {
	if config.MaxConcurrency == 0 {
		config.MaxConcurrency = runtime.NumCPU()
	}
}

// envOverride uses Environment variables to override unspecified properties
func envOverride(config *Config) error {
	if err := envconfig.Process("athens", config); err != nil {
		return err
	}
	return nil
}

func validateConfig(c Config) error {
	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		return err
	}
	return nil
}
