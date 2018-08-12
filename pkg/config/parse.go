package config

import (
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/envconfig"
	validator "gopkg.in/go-playground/validator.v9"
)

// Config provides configuration values for all components
type Config struct {
	EnableCSRFProtection bool           `split_words:"true" default:"false"`
	GoEnv                string         `validate:"required" envconfig:"GO_ENV" default:"development"`
	GoBinary             string         `validate:"required" envconfig:"GO_BINARY_PATH" default:"go"`
	LogLevel             string         `validate:"required" split_words:"true" default:"debug"`
	MaxConcurrency       int            `validate:"required" split_words:"true"`
	MaxWorkerFails       uint           `validate:"required" split_words:"true" default:"5"`
	CloudRuntime         string         `validate:"required" split_words:"true" default:"none"`
	FilterFile           string         `validate:"required" split_words:"true" default:"filter.conf"`
	Timeout              int            `validate:"required" default:"300"`
	Proxy                *ProxyConfig   `validate:""`
	Olympus              *OlympusConfig `validate:""`
	Storage              *StorageConfig `validate:""`
}

// ParseConfig parses the given file into an athens config struct
func ParseConfig(configFile string) (*Config, error) {

	var config Config
	// attempt to parse the given config file
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return nil, err
	}

	// override values with environment variables if specified
	err := envconfig.Process("athens", &config)
	if err != nil {
		return nil, err
	}

	// set default values that are dependent on the runtime
	setRuntimeDefaults(&config)

	// set default timeouts for storage backends if not already set
	setDefaultTimeouts(config.Storage, config.Timeout)

	// validate all required fields have been populated
	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, err
	}
	return &config, nil
}

func setRuntimeDefaults(config *Config) {
	if config.MaxConcurrency == 0 {
		config.MaxConcurrency = runtime.NumCPU()
	}
}
