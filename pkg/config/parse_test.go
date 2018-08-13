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

	envVars := map[string]string{
		"GO_ENV":                        expConf.GoEnv,
		"GO_BINARY_PATH":                expConf.GoBinary,
		"ATHENS_LOG_LEVEL":              expConf.LogLevel,
		"ATHENS_CLOUD_RUNTIME":          expConf.CloudRuntime,
		"ATHENS_MAX_CONCURRENCY":        strconv.Itoa(expConf.MaxConcurrency),
		"ATHENS_MAX_WORKER_FAILS":       strconv.FormatUint(uint64(expConf.MaxWorkerFails), 10),
		"ATHENS_FILTER_FILE":            expConf.FilterFile,
		"ATHENS_TIMEOUT":                strconv.Itoa(expConf.Timeout),
		"ATHENS_ENABLE_CSRF_PROTECTION": strconv.FormatBool(expConf.EnableCSRFProtection),
		"ATHENS_STORAGE_TYPE":           expConf.Proxy.StorageType,
		"OLYMPUS_GLOBAL_ENDPOINT":       expProxy.OlympusGlobalEndpoint,
		"PORT": expProxy.Port,
		"ATHENS_REDIS_QUEUE_PORT":        expProxy.RedisQueueAddress,
		"OLYMPUS_BACKGROUND_WORKER_TYPE": expOlympus.WorkerType,
		"OLYMPUS_REDIS_QUEUE_PORT":       expOlympus.RedisQueueAddress,
	}
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
	envVars := map[string]string{
		"CDN_ENDPOINT":                   expStorage.CDN.Endpoint,
		"ATHENS_DISK_STORAGE_ROOT":       expStorage.Disk.RootPath,
		"GOOGLE_CLOUD_PROJECT":           expStorage.GCP.ProjectID,
		"ATHENS_STORAGE_GCP_BUCKET":      expStorage.GCP.Bucket,
		"ATHENS_MINIO_ENDPOINT":          expStorage.Minio.Endpoint,
		"ATHENS_MINIO_ACCESS_KEY_ID":     expStorage.Minio.Key,
		"ATHENS_MINIO_SECRET_ACCESS_KEY": expStorage.Minio.Secret,
		"ATHENS_MINIO_USE_SSL":           strconv.FormatBool(expStorage.Minio.EnableSSL),
		"ATHENS_MINIO_BUCKET_NAME":       expStorage.Minio.Bucket,
		"ATHENS_MONGO_STORAGE_URL":       expStorage.Mongo.URL,
		"MONGO_CONN_TIMEOUT_SEC":         strconv.Itoa(expStorage.Mongo.Timeout),
		"ATHENS_RDBMS_STORAGE_NAME":      expStorage.RDBMS.Name,
	}
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
	envVars := []string{
		"GO_ENV",
		"GO_BINARY_PATH",
		"ATHENS_LOG_LEVEL",
		"ATHENS_CLOUD_RUNTIME",
		"ATHENS_MAX_CONCURRENCY",
		"ATHENS_MAX_WORKER_FAILS",
		"ATHENS_FILTER_FILE",
		"ATHENS_TIMEOUT",
		"ATHENS_ENABLE_CSRF_PROTECTION",
		"ATHENS_STORAGE_TYPE",
		"OLYMPUS_GLOBAL_ENDPOINT",
		"PORT",
		"ATHENS_REDIS_QUEUE_PORT",
		"OLYMPUS_BACKGROUND_WORKER_TYPE",
		"OLYMPUS_REDIS_QUEUE_PORT",
		"CDN_ENDPOINT",
		"ATHENS_DISK_STORAGE_ROOT",
		"GOOGLE_CLOUD_PROJECT",
		"ATHENS_STORAGE_GCP_BUCKET",
		"ATHENS_MINIO_ENDPOINT",
		"ATHENS_MINIO_ACCESS_KEY_ID",
		"ATHENS_MINIO_SECRET_ACCESS_KEY",
		"ATHENS_MINIO_USE_SSL",
		"ATHENS_MINIO_BUCKET_NAME",
		"ATHENS_MONGO_STORAGE_URL",
		"MONGO_CONN_TIMEOUT_SEC",
		"ATHENS_RDBMS_STORAGE_NAME",
	}
	envVarBackup := map[string]string{}
	for _, k := range envVars {
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

func cleanupEnv(envVars map[string]string) {
	for k, v := range envVars {
		if v != "" {
			os.Setenv(k, v)
		} else {
			os.Unsetenv(k)
		}
	}
}
