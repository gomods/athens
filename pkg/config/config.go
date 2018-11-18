package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/gomods/athens/pkg/errors"
	"github.com/kelseyhightower/envconfig"
	validator "gopkg.in/go-playground/validator.v9"
)

// Config provides configuration values for all components
type Config struct {
	TimeoutConf
	GoEnv            string `validate:"required" envconfig:"GO_ENV"`
	GoBinary         string `validate:"required" envconfig:"GO_BINARY_PATH"`
	GoGetWorkers     int    `validate:"required" envconfig:"ATHENS_GOGET_WORKERS"`
	ProtocolWorkers  int    `validate:"required" envconfig:"ATHENS_PROTOCOL_WORKERS"`
	LogLevel         string `validate:"required" envconfig:"ATHENS_LOG_LEVEL"`
	BuffaloLogLevel  string `validate:"required" envconfig:"BUFFALO_LOG_LEVEL"`
	CloudRuntime     string `validate:"required" envconfig:"ATHENS_CLOUD_RUNTIME"`
	FilterFile       string `envconfig:"ATHENS_FILTER_FILE"`
	TraceExporterURL string `envconfig:"ATHENS_TRACE_EXPORTER_URL"`
	TraceExporter    string `envconfig:"ATHENS_TRACE_EXPORTER"`
	StorageType      string `validate:"required" envconfig:"ATHENS_STORAGE_TYPE"`
	GlobalEndpoint   string `envconfig:"ATHENS_GLOBAL_ENDPOINT"` // This feature is not yet implemented
	Port             string `envconfig:"ATHENS_PORT" default:":3000"`
	BasicAuthUser    string `envconfig:"BASIC_AUTH_USER"`
	BasicAuthPass    string `envconfig:"BASIC_AUTH_PASS"`
	ForceSSL         bool   `envconfig:"PROXY_FORCE_SSL"`
	ValidatorHook    string `envconfig:"ATHENS_PROXY_VALIDATOR"`
	PathPrefix       string `envconfig:"ATHENS_PATH_PREFIX"`
	NETRCPath        string `envconfig:"ATHENS_NETRC_PATH"`
	GithubToken      string `envconfig:"ATHENS_GITHUB_TOKEN"`
	HGRCPath         string `envconfig:"ATHENS_HGRC_PATH"`
	Storage          *StorageConfig
}

// BasicAuth returns BasicAuthUser and BasicAuthPassword
// and ok if neither of them are empty
func (c *Config) BasicAuth() (user, pass string, ok bool) {
	user = c.BasicAuthUser
	pass = c.BasicAuthPass
	ok = user != "" && pass != ""
	return user, pass, ok
}

// FilterOff returns true if the FilterFile is empty
func (c *Config) FilterOff() bool {
	return c.FilterFile == ""
}

// ParseConfigFile parses the given file into an athens config struct
func ParseConfigFile(configFile string) (*Config, error) {

	var config Config
	// attempt to read the given config file
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return nil, err
	}

	// Check file perms from config
	if err := checkFilePerms(config.FilterFile); err != nil {
		return nil, err
	}

	// override values with environment variables if specified
	if err := envOverride(&config); err != nil {
		return nil, err
	}

	// validate all required fields have been populated
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	return &config, nil
}

// envOverride uses Environment variables to override unspecified properties
func envOverride(config *Config) error {
	return envconfig.Process("athens", config)
}

func validateConfig(config Config) error {
	validate := validator.New()
	err := validate.StructExcept(config, "Storage")
	if err != nil {
		return err
	}
	switch config.StorageType {
	case "memory":
		return nil
	case "mongo":
		return validate.Struct(config.Storage.Mongo)
	case "disk":
		return validate.Struct(config.Storage.Disk)
	case "minio":
		return validate.Struct(config.Storage.Minio)
	case "gcp":
		return validate.Struct(config.Storage.GCP)
	case "s3":
		return validate.Struct(config.Storage.S3)
	case "azure":
		return validate.Struct(config.Storage.Azure)
	default:
		return fmt.Errorf("storage type %s is unknown", config.StorageType)
	}
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

// checkFilePerms given a list of files
func checkFilePerms(files ...string) error {
	const op = "config.checkFilePerms"

	for _, f := range files {
		if f == "" {
			continue
		}

		// TODO: Do not ignore errors when a file is not found
		// There is a subtle bug in the filter module which ignores the filter file if it does not find it.
		// This check can be removed once that has been fixed
		fInfo, err := os.Lstat(f)
		if err != nil {
			continue
		}

		if runtime.GOOS == "windows" {
			if fInfo.Mode() != 0400 {
				return errors.E(op, f+" should have 0400 as permission")
			}
		} else {
			// Assume unix based system (MacOS and Linux)
			if fInfo.Mode() != 0640 {
				return errors.E(op, f+" should have 0600 as permission")
			}
		}
	}

	return nil
}
