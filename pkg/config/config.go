package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/errors"
	"github.com/kelseyhightower/envconfig"
)

const defaultConfigFile = "athens.toml"

// Config provides configuration values for all components.
type Config struct {
	TimeoutConf
	GoEnv            string    `envconfig:"GO_ENV"                    validate:"required"`
	GoBinary         string    `envconfig:"GO_BINARY_PATH"            validate:"required"`
	GoBinaryEnvVars  EnvList   `envconfig:"ATHENS_GO_BINARY_ENV_VARS"`
	GoGetWorkers     int       `envconfig:"ATHENS_GOGET_WORKERS"      validate:"required"`
	GoGetDir         string    `envconfig:"ATHENS_GOGET_DIR"`
	ProtocolWorkers  int       `envconfig:"ATHENS_PROTOCOL_WORKERS"   validate:"required"`
	LogLevel         string    `envconfig:"ATHENS_LOG_LEVEL"          validate:"required"`
	LogFormat        string    `envconfig:"ATHENS_LOG_FORMAT"         validate:"oneof='' 'json' 'plain'"`
	CloudRuntime     string    `envconfig:"ATHENS_CLOUD_RUNTIME"      validate:"required_without=LogFormat"`
	EnablePprof      bool      `envconfig:"ATHENS_ENABLE_PPROF"`
	PprofPort        string    `envconfig:"ATHENS_PPROF_PORT"`
	FilterFile       string    `envconfig:"ATHENS_FILTER_FILE"`
	TraceExporterURL string    `envconfig:"ATHENS_TRACE_EXPORTER_URL"`
	TraceExporter    string    `envconfig:"ATHENS_TRACE_EXPORTER"`
	StatsExporter    string    `envconfig:"ATHENS_STATS_EXPORTER"`
	StorageType      string    `envconfig:"ATHENS_STORAGE_TYPE"       validate:"required"`
	GlobalEndpoint   string    `envconfig:"ATHENS_GLOBAL_ENDPOINT"` // This feature is not yet implemented
	Port             string    `envconfig:"ATHENS_PORT"`
	UnixSocket       string    `envconfig:"ATHENS_UNIX_SOCKET"`
	BasicAuthUser    string    `envconfig:"BASIC_AUTH_USER"`
	BasicAuthPass    string    `envconfig:"BASIC_AUTH_PASS"`
	HomeTemplatePath string    `envconfig:"ATHENS_HOME_TEMPLATE_PATH"`
	ForceSSL         bool      `envconfig:"PROXY_FORCE_SSL"`
	ValidatorHook    string    `envconfig:"ATHENS_PROXY_VALIDATOR"`
	PathPrefix       string    `envconfig:"ATHENS_PATH_PREFIX"`
	NETRCPath        string    `envconfig:"ATHENS_NETRC_PATH"`
	GithubToken      string    `envconfig:"ATHENS_GITHUB_TOKEN"`
	HGRCPath         string    `envconfig:"ATHENS_HGRC_PATH"`
	TLSCertFile      string    `envconfig:"ATHENS_TLSCERT_FILE"`
	TLSKeyFile       string    `envconfig:"ATHENS_TLSKEY_FILE"`
	SumDBs           []string  `envconfig:"ATHENS_SUM_DBS"`
	NoSumPatterns    []string  `envconfig:"ATHENS_GONOSUM_PATTERNS"`
	DownloadMode     mode.Mode `envconfig:"ATHENS_DOWNLOAD_MODE"`
	DownloadURL      string    `envconfig:"ATHENS_DOWNLOAD_URL"`
	NetworkMode      string    `envconfig:"ATHENS_NETWORK_MODE"       validate:"oneof=strict offline fallback"`
	SingleFlightType string    `envconfig:"ATHENS_SINGLE_FLIGHT_TYPE"`
	RobotsFile       string    `envconfig:"ATHENS_ROBOTS_FILE"`
	IndexType        string    `envconfig:"ATHENS_INDEX_TYPE"`
	ShutdownTimeout  int       `envconfig:"ATHENS_SHUTDOWN_TIMEOUT"   validate:"min=0"`
	SingleFlight     *SingleFlight
	Storage          *Storage
	Index            *Index
}

// EnvList is a list of key-value environment
// variables that are passed to the Go command.
type EnvList []string

// HasKey returns whether a key-value entry
// is present by only checking the left of
// key=value.
func (el EnvList) HasKey(key string) bool {
	for _, env := range el {
		if strings.HasPrefix(env, key+"=") {
			return true
		}
	}
	return false
}

// Add adds a key=value entry to the environment
// list.
func (el *EnvList) Add(key, value string) {
	*el = append(*el, key+"="+value)
}

// Decode implements envconfig.Decoder. Please see the below link for more information on
// that interface:
//
// https://github.com/kelseyhightower/envconfig#custom-decoders
//
// We are doing this to allow for very long lists of assignments to be set inside of
// a single environment variable. For example:
//
//	ATHENS_GO_BINARY_ENV_VARS="GOPRIVATE=*.corp.example.com,rsc.io/private; GOPROXY=direct"
//
// See the below link for more information:
// https://github.com/gomods/athens/issues/1404
func (el *EnvList) Decode(value string) error {
	const op errors.Op = "envList.Decode"
	if value == "" {
		return nil
	}
	*el = EnvList{} // env vars must override config file
	assignments := strings.Split(value, ";")
	for _, assignment := range assignments {
		*el = append(*el, strings.TrimSpace(assignment))
	}
	return el.Validate()
}

// Validate validates that all strings inside the
// list are of the key=value format.
func (el EnvList) Validate() error {
	const op errors.Op = "EnvList.Validate"
	for _, env := range el {
		// some strings can have multiple "=", such as GODEBUG=netdns=cgo
		if strings.Count(env, "=") < 1 {
			return errors.E(op, fmt.Errorf("incorrect env format: %v", env))
		}
	}
	return nil
}

// Load loads the config from a file.
// If file is not present returns default config.
func Load(configFile string) (*Config, error) {
	// User explicitly specified a config file
	if configFile != "" {
		return ParseConfigFile(configFile)
	}

	// There is a config in the current directory
	if fi, err := os.Stat(defaultConfigFile); err == nil {
		return ParseConfigFile(fi.Name())
	}

	// Use default values
	log.Println("Running dev mode with default settings, consult config when you're ready to run in production")
	cfg := defaultConfig()
	return cfg, envOverride(cfg)
}

func defaultConfig() *Config {
	return &Config{
		GoBinary:         "go",
		GoBinaryEnvVars:  EnvList{"GOPROXY=direct"},
		GoEnv:            "development",
		GoGetWorkers:     10,
		ProtocolWorkers:  30,
		LogLevel:         "debug",
		LogFormat:        "plain",
		CloudRuntime:     "none",
		EnablePprof:      false,
		PprofPort:        ":3001",
		StatsExporter:    "prometheus",
		TimeoutConf:      TimeoutConf{Timeout: 300},
		HomeTemplatePath: "/var/lib/athens/home.html",
		StorageType:      "memory",
		Port:             ":3000",
		SingleFlightType: "memory",
		GlobalEndpoint:   "http://localhost:3001",
		TraceExporterURL: "http://localhost:14268",
		SumDBs:           []string{"https://sum.golang.org"},
		NoSumPatterns:    []string{},
		DownloadMode:     "sync",
		DownloadURL:      "",
		NetworkMode:      "strict",
		RobotsFile:       "robots.txt",
		IndexType:        "none",
		ShutdownTimeout:  60,
		SingleFlight: &SingleFlight{
			Etcd:  &Etcd{"localhost:2379,localhost:22379,localhost:32379"},
			Redis: &Redis{"127.0.0.1:6379", "", DefaultRedisLockConfig()},
			RedisSentinel: &RedisSentinel{
				Endpoints:        []string{"127.0.0.1:26379"},
				MasterName:       "redis-1",
				SentinelPassword: "sekret",
				LockConfig:       DefaultRedisLockConfig(),
			},
			GCP: DefaultGCPConfig(),
		},
		Index: &Index{
			MySQL: &MySQL{
				Protocol: "tcp",
				Host:     "localhost",
				Port:     3306,
				User:     "root",
				Password: "",
				Database: "athens",
				Params: map[string]string{
					"parseTime": "true",
					"timeout":   "30s",
				},
			},
			Postgres: &Postgres{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "",
				Database: "athens",
				Params: map[string]string{
					"connect_timeout": "30",
					"sslmode":         "disable",
				},
			},
		},
	}
}

// BasicAuth returns BasicAuthUser and BasicAuthPassword
// and ok if neither of them are empty.
func (c *Config) BasicAuth() (user, pass string, ok bool) {
	user = c.BasicAuthUser
	pass = c.BasicAuthPass
	ok = user != "" && pass != ""
	return user, pass, ok
}

// FilterOff returns true if the FilterFile is empty.
func (c *Config) FilterOff() bool {
	return c.FilterFile == ""
}

// ParseConfigFile parses the given file into an athens config struct.
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

	// Check file perms from config
	if config.GoEnv == "production" {
		if err := checkFilePerms(configFile, config.FilterFile); err != nil {
			return nil, err
		}
	}

	// validate all required fields have been populated
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	return &config, nil
}

// envOverride uses Environment variables to override unspecified properties.
func envOverride(config *Config) error {
	const defaultPort = ":3000"
	err := envconfig.Process("athens", config)
	if err != nil {
		return err
	}
	portEnv := os.Getenv("PORT")
	// ATHENS_PORT takes precedence over PORT
	if portEnv != "" && os.Getenv("ATHENS_PORT") == "" {
		config.Port = portEnv
	}
	if config.Port == "" {
		config.Port = defaultPort
	}
	config.Port = ensurePortFormat(config.Port)
	return nil
}

func ensurePortFormat(s string) string {
	if _, err := strconv.Atoi(s); err == nil {
		return ":" + s
	}
	return s
}

func validateConfig(config Config) error {
	validate := validator.New()
	err := validate.StructExcept(config, "Storage", "Index")
	if err != nil {
		return err
	}
	err = validateStorage(validate, config.StorageType, config.Storage)
	if err != nil {
		return err
	}
	err = validateIndex(validate, config.IndexType, config.Index)
	if err != nil {
		return err
	}
	return nil
}

func validateStorage(validate *validator.Validate, storageType string, config *Storage) error {
	switch storageType {
	case "memory":
		return nil
	case "mongo":
		return validate.Struct(config.Mongo)
	case "disk":
		return validate.Struct(config.Disk)
	case "minio":
		return validate.Struct(config.Minio)
	case "gcp":
		return validate.Struct(config.GCP)
	case "s3":
		return validate.Struct(config.S3)
	case "azureblob":
		return validate.Struct(config.AzureBlob)
	case "external":
		return validate.Struct(config.External)
	default:
		return fmt.Errorf("storage type %q is unknown", storageType)
	}
}

func validateIndex(validate *validator.Validate, indexType string, config *Index) error {
	switch indexType {
	case "", "none", "memory":
		return nil
	case "mysql":
		return validate.Struct(config.MySQL)
	case "postgres":
		return validate.Struct(config.Postgres)
	default:
		return fmt.Errorf("index type %q is unknown", indexType)
	}
}

// GetConf accepts the path to a file, constructs an absolute path to the file,
// and attempts to parse it into a Config struct.
func GetConf(path string) (*Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("unable to construct absolute path to test config file")
	}
	conf, err := ParseConfigFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("unable to parse test config file: %w", err)
	}
	return conf, nil
}

// checkFilePerms given a list of files.
func checkFilePerms(files ...string) error {
	const op = "config.checkFilePerms"

	for _, f := range files {
		if f == "" {
			continue
		}

		// TODO: Do not ignore errors when a file is not found
		// There is a subtle bug in the filter module which ignores the filter file if it does not find it.
		// This check can be removed once that has been fixed
		fInfo, err := os.Stat(f)
		if err != nil {
			continue
		}

		// Assume unix based system (MacOS and Linux)
		// the bit mask is calculated using the umask command which tells which permissions
		// should not be allowed for a particular user, group or world
		if fInfo.Mode()&0o033 != 0 && runtime.GOOS != "windows" {
			return errors.E(op, f+" should have at most rwx,-, - (bit mask 077) as permission")
		}
	}

	return nil
}
