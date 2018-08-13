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

	// Set values that are not defaults for everything
	expProxy := ProxyConfig{
		StorageType:           "minio",
		OlympusGlobalEndpoint: "mytikas.gomods.io",
		RedisQueueAddress:     ":6380",
		Port:                  ":7000",
	}

	expOlympus := OlympusConfig{
		StorageType:       "minio",
		RedisQueueAddress: ":6381",
		Port:              ":7000",
		WorkerType:        "memory",
	}

	expConf := &Config{
		GoEnv:                "production",
		LogLevel:             "info",
		GoBinary:             "go11",
		MaxConcurrency:       4,
		MaxWorkerFails:       10,
		CloudRuntime:         "gcp",
		FilterFile:           "filter2.conf",
		Timeout:              30,
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
	cleanupEnv(envVarBackup)
}

func TestStorageEnvOverrides(t *testing.T) {

	globalTimeout := 300

	// Set values that are not defaults for everything
	expStorage := &StorageConfig{
		CDN: &CDNConfig{
			Endpoint: "cdnEndpoint",
			Timeout:  globalTimeout,
		},
		Disk: &DiskConfig{
			RootPath: "/my/root/path",
		},
		GCP: &GCPConfig{
			ProjectID: "gcpproject",
			Bucket:    "gcpbucket",
			Timeout:   globalTimeout,
		},
		Minio: &MinioConfig{
			Endpoint:  "minioEndpoint",
			Key:       "minioKey",
			Secret:    "minioSecret",
			EnableSSL: false,
			Bucket:    "minioBucket",
			Timeout:   globalTimeout,
		},
		Mongo: &MongoConfig{
			URL:     "mongoURL",
			Timeout: 25,
		},
		RDBMS: &RDBMSConfig{
			Name:    "production",
			Timeout: globalTimeout,
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
	setDefaultTimeouts(conf.Storage, globalTimeout)
	deleteInvalidStorageConfigs(conf.Storage)

	eq := cmp.Equal(conf.Storage, expStorage)
	if !eq {
		t.Error("Environment variables did not correctly override storage config values")
	}
	cleanupEnv(envVarBackup)
}

func TestParseDefaultsSuccess(t *testing.T) {
	_, err := parseConfig("")
	if err != nil {
		t.Errorf("Default values are causing validation failures")
	}
}

// TestParseExampleConfig validates that all the properties in the example configuration file
// can be parsed and validated without any environment variables
func TestParseExampleConfig(t *testing.T) {

	// unset all applicable environment variables
	envVars := getEnvMap(&Config{})
	envVarBackup := map[string]string{}
	for k := range envVars {
		oldVal := os.Getenv(k)
		envVarBackup[k] = oldVal
		os.Unsetenv(k)
	}

	globalTimeout := 300

	expProxy := &ProxyConfig{
		StorageType:           "mongo",
		OlympusGlobalEndpoint: "olympus.gomods.io",
		RedisQueueAddress:     ":6379",
		Port:                  ":3000",
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
			Timeout:  globalTimeout,
		},
		Disk: &DiskConfig{
			RootPath: "/path/on/disk",
		},
		GCP: &GCPConfig{
			ProjectID: "MY_GCP_PROJECT_ID",
			Bucket:    "MY_GCP_BUCKET",
			Timeout:   globalTimeout,
		},
		Minio: &MinioConfig{
			Endpoint:  "minio.example.com",
			Key:       "MY_KEY",
			Secret:    "MY_SECRET",
			EnableSSL: true,
			Bucket:    "gomods",
			Timeout:   globalTimeout,
		},
		Mongo: &MongoConfig{
			URL:     "mongo.example.com",
			Timeout: globalTimeout,
		},
		RDBMS: &RDBMSConfig{
			Name:    "development",
			Timeout: globalTimeout,
		},
	}

	expConf := &Config{
		GoEnv:                "development",
		LogLevel:             "debug",
		GoBinary:             "go",
		MaxConcurrency:       4,
		MaxWorkerFails:       5,
		CloudRuntime:         "none",
		FilterFile:           "filter.conf",
		Timeout:              300,
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
	cleanupEnv(envVarBackup)
}

func getEnvMap(config *Config) map[string]string {

	envVars := map[string]string{
		"GO_ENV":                        config.GoEnv,
		"GO_BINARY_PATH":                config.GoBinary,
		"ATHENS_LOG_LEVEL":              config.LogLevel,
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
		if storage.RDBMS != nil {
			envVars["ATHENS_RDBMS_STORAGE_NAME"] = storage.RDBMS.Name
		}
	}
	return envVars
}

func cleanupEnv(envVars map[string]string) {
	for k, v := range envVars {
		if v != "" {
			os.Setenv(k, v)
		} else {
			os.Unsetenv(k)
		}
	}
}
