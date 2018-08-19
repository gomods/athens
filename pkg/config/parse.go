package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/envconfig"
	validator "gopkg.in/go-playground/validator.v9"
)

// Config provides configuration values for all components
type Config struct {
	GoEnv                string         `validate:"required" envconfig:"GO_ENV"`
	GoBinary             string         `validate:"required" envconfig:"GO_BINARY_PATH"`
	LogLevel             string         `validate:"required" split_words:"true"`
	MaxConcurrency       int            `validate:"required" split_words:"true"`
	MaxWorkerFails       uint           `validate:"required" split_words:"true"`
	CloudRuntime         string         `validate:"required" split_words:"true"`
	FilterFile           string         `validate:"required" split_words:"true"`
	Timeout              int            `validate:"required"`
	EnableCSRFProtection bool           `envconfig:"ATHENS_ENABLE_CSRF_PROTECTION"`
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

	// set default values
	setDefaults(&config)

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

func setDefaults(config *Config) {
	if config.MaxConcurrency == 0 {
		config.MaxConcurrency = runtime.NumCPU()
	}

	overrideDefaultStr(&config.GoEnv, "development")
	overrideDefaultStr(&config.GoBinary, "go")
	overrideDefaultStr(&config.LogLevel, "debug")
	overrideDefaultStr(&config.CloudRuntime, "none")
	overrideDefaultStr(&config.FilterFile, "filter.conf")
	overrideDefaultInt(&config.Timeout, 300)
	overrideDefaultUint(&config.MaxWorkerFails, 5)

	config.Proxy = setProxyDefaults(config.Proxy)
	config.Olympus = setOlympusDefaults(config.Olympus)
	config.Storage = setStorageDefaults(config.Storage, config.Timeout)
}

func overrideDefaultUint(val *uint, override uint) {
	if *val == 0 {
		*val = override
	}
}

func overrideDefaultInt(val *int, override int) {
	if *val == 0 {
		*val = override
	}
}

func overrideDefaultStr(val *string, override string) {
	if *val == "" {
		*val = override
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
	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		return err
	}
	return nil
}

// GetConf accepts the path to a file, constructs an absolute path to the file,
// and attempts to parse it into a Config struct.
func GetConf(path string) (*Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to construct absolute path to test config file")
	}
	conf, err := ParseConfigFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse test config file: %s", err.Error())
	}
	return conf, nil
}

// GetConfLogErr is similar to GetConf, except it logs a failure for the calling test
// if any errors are encountered
func GetConfLogErr(path string, t *testing.T) *Config {
	c, err := GetConf(path)
	if err != nil {
		t.Fatalf("Unable to parse config file: %s", err.Error())
	}
	return c
}
