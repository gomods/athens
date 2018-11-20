package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const exampleConfigPath = "../../config.dev.toml"

func compareConfigs(parsedConf *Config, expConf *Config, t *testing.T) {
	opts := cmpopts.IgnoreTypes(StorageConfig{})
	eq := cmp.Equal(parsedConf, expConf, opts)
	if !eq {
		t.Errorf("Parsed Example configuration did not match expected values. Expected: %+v. Actual: %+v", expConf, parsedConf)
	}
}

func compareStorageConfigs(parsedStorage *StorageConfig, expStorage *StorageConfig, t *testing.T) {
	eq := cmp.Equal(parsedStorage.Mongo, expStorage.Mongo)
	if !eq {
		t.Errorf("Parsed Example Storage configuration did not match expected values. Expected: %+v. Actual: %+v", expStorage.Mongo, parsedStorage.Mongo)
	}
	eq = cmp.Equal(parsedStorage.Minio, expStorage.Minio)
	if !eq {
		t.Errorf("Parsed Example Storage configuration did not match expected values. Expected: %+v. Actual: %+v", expStorage.Minio, parsedStorage.Minio)
	}
	eq = cmp.Equal(parsedStorage.Disk, expStorage.Disk)
	if !eq {
		t.Errorf("Parsed Example Storage configuration did not match expected values. Expected: %+v. Actual: %+v", expStorage.Disk, parsedStorage.Disk)
	}
	eq = cmp.Equal(parsedStorage.GCP, expStorage.GCP)
	if !eq {
		t.Errorf("Parsed Example Storage configuration did not match expected values. Expected: %+v. Actual: %+v", expStorage.GCP, parsedStorage.GCP)
	}
	eq = cmp.Equal(parsedStorage.S3, expStorage.S3)
	if !eq {
		t.Errorf("Parsed Example Storage configuration did not match expected values. Expected: %+v. Actual: %+v", expStorage.S3, parsedStorage.S3)
	}
}

func TestPortDefaultsCorrectly(t *testing.T) {
	conf := &Config{}
	err := envOverride(conf)
	if err != nil {
		t.Fatalf("Env override failed: %v", err)
	}
	expPort := ":3000"
	if conf.Port != expPort {
		t.Errorf("Port was incorrect. Got: %s, want: %s", conf.Port, expPort)
	}
}

func TestEnvOverrides(t *testing.T) {
	expConf := &Config{
		GoEnv:           "production",
		GoGetWorkers:    10,
		ProtocolWorkers: 10,
		LogLevel:        "info",
		BuffaloLogLevel: "info",
		GoBinary:        "go11",
		CloudRuntime:    "gcp",
		TimeoutConf: TimeoutConf{
			Timeout: 30,
		},
		StorageType:    "minio",
		GlobalEndpoint: "mytikas.gomods.io",
		Port:           ":7000",
		BasicAuthUser:  "testuser",
		BasicAuthPass:  "testpass",
		ForceSSL:       true,
		ValidatorHook:  "testhook.io",
		PathPrefix:     "prefix",
		NETRCPath:      "/test/path/.netrc",
		HGRCPath:       "/test/path/.hgrc",
		Storage:        &StorageConfig{},
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
	compareConfigs(conf, expConf, t)
	restoreEnv(envVarBackup)
}

func TestStorageEnvOverrides(t *testing.T) {
	expStorage := &StorageConfig{
		Disk: &DiskConfig{
			RootPath: "/my/root/path",
		},
		GCP: &GCPConfig{
			ProjectID: "gcpproject",
			Bucket:    "gcpbucket",
		},
		Minio: &MinioConfig{
			Endpoint:  "minioEndpoint",
			Key:       "minioKey",
			Secret:    "minioSecret",
			EnableSSL: false,
			Bucket:    "minioBucket",
			Region:    "us-west-1",
		},
		Mongo: &MongoConfig{
			URL:      "mongoURL",
			CertPath: "/test/path",
		},
		S3: &S3Config{
			Region: "s3Region",
			Key:    "s3Key",
			Secret: "s3Secret",
			Token:  "s3Token",
			Bucket: "s3Bucket",
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
	compareStorageConfigs(conf.Storage, expStorage, t)
	restoreEnv(envVarBackup)
}

// TestParseExampleConfig validates that all the properties in the example configuration file
// can be parsed and validated without any environment variables
func TestParseExampleConfig(t *testing.T) {

	// initialize all struct pointers so we get all applicable env variables
	emptyConf := &Config{
		Storage: &StorageConfig{
			Disk: &DiskConfig{},
			GCP:  &GCPConfig{},
			Minio: &MinioConfig{
				EnableSSL: false,
			},
			Mongo: &MongoConfig{},
			S3:    &S3Config{},
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

	expStorage := &StorageConfig{
		Disk: &DiskConfig{
			RootPath: "/path/on/disk",
		},
		GCP: &GCPConfig{
			ProjectID: "MY_GCP_PROJECT_ID",
			Bucket:    "MY_GCP_BUCKET",
		},
		Minio: &MinioConfig{
			Endpoint:  "127.0.0.1:9001",
			Key:       "minio",
			Secret:    "minio123",
			EnableSSL: false,
			Bucket:    "gomods",
		},
		Mongo: &MongoConfig{
			URL:          "mongodb://127.0.0.1:27017",
			CertPath:     "",
			InsecureConn: false,
		},
		S3: &S3Config{
			Region: "MY_AWS_REGION",
			Key:    "MY_AWS_ACCESS_KEY_ID",
			Secret: "MY_AWS_SECRET_ACCESS_KEY",
			Token:  "",
			Bucket: "MY_S3_BUCKET_NAME",
		},
	}

	expConf := &Config{
		GoEnv:           "development",
		LogLevel:        "debug",
		BuffaloLogLevel: "debug",
		GoBinary:        "go",
		GoGetWorkers:    30,
		ProtocolWorkers: 30,
		CloudRuntime:    "none",
		TimeoutConf: TimeoutConf{
			Timeout: 300,
		},
		StorageType:    "memory",
		GlobalEndpoint: "http://localhost:3001",
		Port:           ":3000",
		BasicAuthUser:  "",
		BasicAuthPass:  "",
		Storage:        expStorage,
	}

	absPath, err := filepath.Abs(exampleConfigPath)
	if err != nil {
		t.Errorf("Unable to construct absolute path to example config file")
	}
	parsedConf, err := ParseConfigFile(absPath)
	if err != nil {
		t.Errorf("Unable to parse example config file: %+v", err)
	}

	compareConfigs(parsedConf, expConf, t)
	restoreEnv(envVarBackup)
}

func getEnvMap(config *Config) map[string]string {

	envVars := map[string]string{
		"GO_ENV":                  config.GoEnv,
		"GO_BINARY_PATH":          config.GoBinary,
		"ATHENS_GOGET_WORKERS":    strconv.Itoa(config.GoGetWorkers),
		"ATHENS_PROTOCOL_WORKERS": strconv.Itoa(config.ProtocolWorkers),
		"ATHENS_LOG_LEVEL":        config.LogLevel,
		"BUFFALO_LOG_LEVEL":       config.BuffaloLogLevel,
		"ATHENS_CLOUD_RUNTIME":    config.CloudRuntime,
		"ATHENS_TIMEOUT":          strconv.Itoa(config.Timeout),
		"ATHENS_TRACE_EXPORTER":   config.TraceExporterURL,
	}

	envVars["ATHENS_STORAGE_TYPE"] = config.StorageType
	envVars["ATHENS_GLOBAL_ENDPOINT"] = config.GlobalEndpoint
	envVars["ATHENS_PORT"] = config.Port
	envVars["BASIC_AUTH_USER"] = config.BasicAuthUser
	envVars["BASIC_AUTH_PASS"] = config.BasicAuthPass
	envVars["PROXY_FORCE_SSL"] = strconv.FormatBool(config.ForceSSL)
	envVars["ATHENS_PROXY_VALIDATOR"] = config.ValidatorHook
	envVars["ATHENS_PATH_PREFIX"] = config.PathPrefix
	envVars["ATHENS_NETRC_PATH"] = config.NETRCPath
	envVars["ATHENS_HGRC_PATH"] = config.HGRCPath

	storage := config.Storage
	if storage != nil {
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
			envVars["ATHENS_MINIO_REGION"] = storage.Minio.Region
			envVars["ATHENS_MINIO_BUCKET_NAME"] = storage.Minio.Bucket
		}
		if storage.Mongo != nil {
			envVars["ATHENS_MONGO_STORAGE_URL"] = storage.Mongo.URL
			envVars["ATHENS_MONGO_CERT_PATH"] = storage.Mongo.CertPath
			envVars["ATHENS_MONGO_INSECURE"] = strconv.FormatBool(storage.Mongo.InsecureConn)
		}
		if storage.S3 != nil {
			envVars["AWS_REGION"] = storage.S3.Region
			envVars["AWS_ACCESS_KEY_ID"] = storage.S3.Key
			envVars["AWS_SECRET_ACCESS_KEY"] = storage.S3.Secret
			envVars["AWS_SESSION_TOKEN"] = storage.S3.Token
			envVars["ATHENS_S3_BUCKET_NAME"] = storage.S3.Bucket
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

func invalidPerm() os.FileMode {
	if runtime.GOOS == "windows" {
		return 0200
	}
	return 0777
}

func correctPerm() os.FileMode {
	if runtime.GOOS == "windows" {
		return 0600
	}
	return 0640
}

func Test_checkFilePerms(t *testing.T) {
	f1, err := ioutil.TempFile(os.TempDir(), "prefix-")
	if err != nil {
		t.FailNow()
	}
	defer os.Remove(f1.Name())
	err = os.Chmod(f1.Name(), invalidPerm())

	// t.Logf("for windows invalidPerm for f1 is %s is %d and error is %s\n", f1.Name(), invalidPerm(), err)
	stat, lstatErr := os.Lstat(f1.Name())
	t.Logf("f1 stat: %d, err %s", stat.Mode(), lstatErr)

	stat, lstatErr := os.Lstat(f2.Name())
	t.Logf("f2 stat: %d, err %s", stat.Mode(), lstatErr)

	f2, err := ioutil.TempFile(os.TempDir(), "prefix-")
	if err != nil {
		t.FailNow()
	}
	defer os.Remove(f2.Name())
	err = os.Chmod(f2.Name(), correctPerm())

	type args struct {
		files []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"should not have an error on empty file name",
			args{
				[]string{"", f2.Name()},
			},
			false,
		},
		{
			"should have an error if all the files have incorrect permissions",
			args{
				[]string{f1.Name(), f1.Name(), f1.Name()},
			},
			true,
		},
		{
			"should have an error when at least 1 file has wrong permissions",
			args{
				[]string{f2.Name(), f1.Name()},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkFilePerms(tt.args.files...); (err != nil) != tt.wantErr {
				t.Errorf("checkFilePerms() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
