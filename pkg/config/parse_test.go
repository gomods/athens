package config

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const exampleConfigPath = "../../config.example.toml"

func TestEnvOverrides(t *testing.T) {

	expProxy := ProxyConfig{
		StorageType:           "minio",
		OlympusGlobalEndpoint: "mytikas.gomods.io",
		RedisQueueAddress:     ":6380",
		Port:                  ":7000",
		FilterOff:             false,
		BasicAuthUser:         "testuser",
		BasicAuthPass:         "testpass",
		ForceSSL:              true,
		ValidatorHook:         "testhook.io",
		PathPrefix:            "prefix",
		NETRCPath:             "/test/path",
	}

	expOlympus := OlympusConfig{
		StorageType:       "minio",
		RedisQueueAddress: ":6381",
		Port:              ":7000",
		WorkerType:        "memory",
	}

	expConf := &Config{
		GoEnv:           "production",
		GoGetWorkers:    10,
		ProtocolWorkers: 10,
		LogLevel:        "info",
		BuffaloLogLevel: "info",
		GoBinary:        "go11",
		MaxConcurrency:  4,
		MaxWorkerFails:  10,
		CloudRuntime:    "gcp",
		FilterFile:      "filter2.conf",
		TimeoutConf: TimeoutConf{
			Timeout: 30,
		},
		EnableCSRFProtection: true,
		Proxy:                &expProxy,
		Olympus:              &expOlympus,
		Storage:              &StorageConfig{},
	}

	envVars := getEnvMap(expConf)
	envVarBackup := map[string]string{}
	for k, v := range envVars {
		oldVal := os.Getenv(k)
		envVarBackup[k] = oldVal
		os.Setenv(k, v)
	}
	conf := &Config{}
	err := envOverride(conf)
	if err != nil {
		t.Fatalf("Env override failed: %v", err)
	}
	deleteInvalidStorageConfigs(conf.Storage)

	eq := cmp.Equal(conf, expConf)
	if !eq {
		t.Errorf("Environment variables did not correctly override config values. Expected: %+v. Actual: %+v", expConf, conf)
	}
	restoreEnv(envVarBackup)
}

func TestStorageEnvOverrides(t *testing.T) {

	globalTimeout := 300
	expStorage := &StorageConfig{
		CDN: &CDNConfig{
			Endpoint: "cdnEndpoint",
			TimeoutConf: TimeoutConf{
				Timeout: globalTimeout,
			},
		},
		Disk: &DiskConfig{
			RootPath: "/my/root/path",
		},
		GCP: &GCPConfig{
			ProjectID: "gcpproject",
			Bucket:    "gcpbucket",
			TimeoutConf: TimeoutConf{
				Timeout: globalTimeout,
			},
		},
		Minio: &MinioConfig{
			Endpoint:  "minioEndpoint",
			Key:       "minioKey",
			Secret:    "minioSecret",
			EnableSSL: false,
			Bucket:    "minioBucket",
			TimeoutConf: TimeoutConf{
				Timeout: globalTimeout,
			},
		},
		Mongo: &MongoConfig{
			URL: "mongoURL",
			TimeoutConf: TimeoutConf{
				Timeout: globalTimeout,
			},
		},
	}
	envVars := getEnvMap(&Config{Storage: expStorage})
	envVarBackup := map[string]string{}
	for k, v := range envVars {
		oldVal := os.Getenv(k)
		envVarBackup[k] = oldVal
		os.Setenv(k, v)
	}
	conf := &Config{}
	err := envOverride(conf)
	if err != nil {
		t.Fatalf("Env override failed: %v", err)
	}
	setStorageTimeouts(conf.Storage, globalTimeout)
	deleteInvalidStorageConfigs(conf.Storage)

	eq := cmp.Equal(conf.Storage, expStorage)
	if !eq {
		t.Error("Environment variables did not correctly override storage config values")
	}
	restoreEnv(envVarBackup)
}

// TestParseExampleConfig validates that all the properties in the example configuration file
// can be parsed and validated without any environment variables
func TestParseExampleConfig(t *testing.T) {

	// initialize all struct pointers so we get all applicable env variables
	emptyConf := &Config{
		Proxy:   &ProxyConfig{},
		Olympus: &OlympusConfig{},
		Storage: &StorageConfig{
			CDN:  &CDNConfig{},
			Disk: &DiskConfig{},
			GCP:  &GCPConfig{},
			Minio: &MinioConfig{
				EnableSSL: false,
			},
			Mongo: &MongoConfig{},
		},
	}
	// unset all environment variables
	envVars := getEnvMap(emptyConf)
	envVarBackup := map[string]string{}
	for k := range envVars {
		oldVal := os.Getenv(k)
		envVarBackup[k] = oldVal
		os.Unsetenv(k)
	}

	globalTimeout := 300

	expProxy := &ProxyConfig{
		StorageType:           "memory",
		OlympusGlobalEndpoint: "http://localhost:3001",
		RedisQueueAddress:     ":6379",
		Port:                  ":3000",
		FilterOff:             true,
		BasicAuthUser:         "",
		BasicAuthPass:         "",
	}

	expOlympus := &OlympusConfig{
		StorageType:       "memory",
		RedisQueueAddress: ":6379",
		Port:              ":3001",
		WorkerType:        "redis",
	}

	expStorage := &StorageConfig{
		CDN: &CDNConfig{
			Endpoint: "cdn.example.com",
			TimeoutConf: TimeoutConf{
				Timeout: globalTimeout,
			},
		},
		Disk: &DiskConfig{
			RootPath: "/path/on/disk",
		},
		GCP: &GCPConfig{
			ProjectID: "MY_GCP_PROJECT_ID",
			Bucket:    "MY_GCP_BUCKET",
			TimeoutConf: TimeoutConf{
				Timeout: globalTimeout,
			},
		},
		Minio: &MinioConfig{
			Endpoint:  "minio.example.com",
			Key:       "MY_KEY",
			Secret:    "MY_SECRET",
			EnableSSL: true,
			Bucket:    "gomods",
			TimeoutConf: TimeoutConf{
				Timeout: globalTimeout,
			},
		},
		Mongo: &MongoConfig{
			URL: "mongo.example.com",
			TimeoutConf: TimeoutConf{
				Timeout: globalTimeout,
			},
		},
	}

	expConf := &Config{
		GoEnv:           "development",
		LogLevel:        "debug",
		BuffaloLogLevel: "debug",
		GoBinary:        "go",
		GoGetWorkers:    30,
		ProtocolWorkers: 30,
		MaxConcurrency:  4,
		MaxWorkerFails:  5,
		CloudRuntime:    "none",
		FilterFile:      "filter.conf",
		TimeoutConf: TimeoutConf{
			Timeout: 300,
		},
		EnableCSRFProtection: false,
		Proxy:                expProxy,
		Olympus:              expOlympus,
		Storage:              expStorage,
	}

	absPath, err := filepath.Abs(exampleConfigPath)
	if err != nil {
		t.Errorf("Unable to construct absolute path to example config file")
	}
	parsedConf, err := ParseConfigFile(absPath)
	if err != nil {
		t.Errorf("Unable to parse example config file: %+v", err)
	}

	eq := cmp.Equal(parsedConf, expConf)
	if !eq {
		t.Errorf("Parsed Example configuration did not match expected values. Expected: %+v. Actual: %+v", expConf, parsedConf)
	}
	restoreEnv(envVarBackup)
}

// TestConfigOverridesDefault validates that a value provided by the config is not overwritten during parsing
func TestConfigOverridesDefault(t *testing.T) {

	// set values to anything but defaults
	config := &Config{
		TimeoutConf: TimeoutConf{
			Timeout: 1,
		},
		Storage: &StorageConfig{
			Minio: &MinioConfig{
				Bucket:    "notgomods",
				EnableSSL: false,
				TimeoutConf: TimeoutConf{
					Timeout: 42,
				},
			},
		},
	}

	// should be identical to config above
	expConfig := &Config{
		TimeoutConf: config.TimeoutConf,
		Storage: &StorageConfig{
			Minio: &MinioConfig{
				Bucket:      config.Storage.Minio.Bucket,
				EnableSSL:   config.Storage.Minio.EnableSSL,
				TimeoutConf: config.Storage.Minio.TimeoutConf,
			},
		},
	}

	// unset all environment variables
	envVars := getEnvMap(&Config{})
	envVarBackup := map[string]string{}
	for k := range envVars {
		oldVal := os.Getenv(k)
		envVarBackup[k] = oldVal
		os.Unsetenv(k)
	}

	envOverride(config)

	if config.Timeout != expConfig.Timeout {
		t.Errorf("Default timeout is overriding specified timeout")
	}

	if !cmp.Equal(config.Storage.Minio, expConfig.Storage.Minio) {
		t.Errorf("Default Minio config is overriding specified config")
	}

	restoreEnv(envVarBackup)
}

func getEnvMap(config *Config) map[string]string {

	envVars := map[string]string{
		"GO_ENV":                        config.GoEnv,
		"GO_BINARY_PATH":                config.GoBinary,
		"ATHENS_GOGET_WORKERS":          strconv.Itoa(config.GoGetWorkers),
		"ATHENS_PROTOCOL_WORKERS":       strconv.Itoa(config.ProtocolWorkers),
		"ATHENS_LOG_LEVEL":              config.LogLevel,
		"BUFFALO_LOG_LEVEL":             config.BuffaloLogLevel,
		"ATHENS_CLOUD_RUNTIME":          config.CloudRuntime,
		"ATHENS_MAX_CONCURRENCY":        strconv.Itoa(config.MaxConcurrency),
		"ATHENS_MAX_WORKER_FAILS":       strconv.FormatUint(uint64(config.MaxWorkerFails), 10),
		"ATHENS_FILTER_FILE":            config.FilterFile,
		"ATHENS_TIMEOUT":                strconv.Itoa(config.Timeout),
		"ATHENS_ENABLE_CSRF_PROTECTION": strconv.FormatBool(config.EnableCSRFProtection),
	}

	proxy := config.Proxy
	if proxy != nil {
		envVars["ATHENS_STORAGE_TYPE"] = proxy.StorageType
		envVars["OLYMPUS_GLOBAL_ENDPOINT"] = proxy.OlympusGlobalEndpoint
		envVars["PORT"] = proxy.Port
		envVars["ATHENS_REDIS_QUEUE_PORT"] = proxy.RedisQueueAddress
		envVars["PROXY_FILTER_OFF"] = strconv.FormatBool(proxy.FilterOff)
		envVars["BASIC_AUTH_USER"] = proxy.BasicAuthUser
		envVars["BASIC_AUTH_PASS"] = proxy.BasicAuthPass
		envVars["PROXY_FORCE_SSL"] = strconv.FormatBool(proxy.ForceSSL)
		envVars["ATHENS_PROXY_VALIDATOR"] = proxy.ValidatorHook
		envVars["ATHENS_PATH_PREFIX"] = proxy.PathPrefix
		envVars["ATHENS_NETRC_PATH"] = proxy.NETRCPath
	}

	olympus := config.Olympus
	if olympus != nil {
		envVars["OLYMPUS_BACKGROUND_WORKER_TYPE"] = olympus.WorkerType
		envVars["OLYMPUS_REDIS_QUEUE_PORT"] = olympus.RedisQueueAddress
	}

	storage := config.Storage
	if storage != nil {
		if storage.CDN != nil {
			envVars["CDN_ENDPOINT"] = storage.CDN.Endpoint
		}
		if storage.Disk != nil {
			envVars["ATHENS_DISK_STORAGE_ROOT"] = storage.Disk.RootPath
		}
		if storage.GCP != nil {
			envVars["GOOGLE_CLOUD_PROJECT"] = storage.GCP.ProjectID
			envVars["ATHENS_STORAGE_GCP_BUCKET"] = storage.GCP.Bucket
		}
		if storage.Minio != nil {
			envVars["ATHENS_MINIO_ENDPOINT"] = storage.Minio.Endpoint
			envVars["ATHENS_MINIO_ACCESS_KEY_ID"] = storage.Minio.Key
			envVars["ATHENS_MINIO_SECRET_ACCESS_KEY"] = storage.Minio.Secret
			envVars["ATHENS_MINIO_USE_SSL"] = strconv.FormatBool(storage.Minio.EnableSSL)
			envVars["ATHENS_MINIO_BUCKET_NAME"] = storage.Minio.Bucket
		}
		if storage.Mongo != nil {
			envVars["ATHENS_MONGO_STORAGE_URL"] = storage.Mongo.URL
			envVars["MONGO_CONN_TIMEOUT_SEC"] = strconv.Itoa(storage.Mongo.Timeout)
		}
	}
	return envVars
}

func restoreEnv(envVars map[string]string) {
	for k, v := range envVars {
		if v != "" {
			os.Setenv(k, v)
		} else {
			os.Unsetenv(k)
		}
	}
}
