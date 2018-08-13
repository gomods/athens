package config

import (
	"io/ioutil"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/envconfig"
	validator "gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

func getValidator() *validator.Validate {
	if validate == nil {
		return validator.New()
	}
	return validate
}

// Config provides configuration values for all components
type Config struct {
	GoEnv                string         `validate:"required" envconfig:"GO_ENV" default:"development"`
	GoBinary             string         `validate:"required" envconfig:"GO_BINARY_PATH" default:"go"`
	LogLevel             string         `validate:"required" split_words:"true" default:"debug"`
	MaxConcurrency       int            `validate:"required" split_words:"true"`
	MaxWorkerFails       uint           `validate:"required" split_words:"true" default:"5"`
	CloudRuntime         string         `validate:"required" split_words:"true" default:"none"`
	FilterFile           string         `validate:"required" split_words:"true" default:"filter.conf"`
	Timeout              int            `validate:"required" default:"300"`
	EnableCSRFProtection bool           `envconfig:"ATHENS_ENABLE_CSRF_PROTECTION" default:"false"`
	Proxy                *ProxyConfig   `validate:""`
	Olympus              *OlympusConfig `validate:""`
	Storage              *StorageConfig `validate:""`
}

// ParseConfigFile parses the given file into an athens config struct
func ParseConfigFile(configFile string) (*Config, error) {

	// attempt to read the given config file
	confBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	confStr := string(confBytes)
	return parseConfig(confStr)
}

func parseConfig(confStr string) (*Config, error) {
	var config Config
	if _, err := toml.Decode(confStr, &config); err != nil {
		return nil, err
	}

	// override values with environment variables if specified
	if err := envOverride(&config); err != nil {
		return nil, err
	}

	// set default values that are dependent on the runtime
	setRuntimeDefaults(&config)

	// set default timeouts for storage backends if not already set
	setDefaultTimeouts(config.Storage, config.Timeout)

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

// envOverride uses Environment variables and specified defaults to override unspecified properties
func envOverride(config *Config) error {
	if err := envconfig.Process("athens", config); err != nil {
		return err
	}
	return nil
}

func validateConfig(c Config) error {
	validate := getValidator()
	err := validate.Struct(c)
	if err != nil {
		return err
	}
	return nil
}
