package actions

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/gomods/athens/pkg/config"
)

var testStorage = &config.Storage{
	Disk: &config.DiskConfig{
		RootPath: "/path/on/disk",
	},
	GCP: &config.GCPConfig{
		ProjectID: "MY_GCP_PROJECT_ID",
		Bucket:    "MY_GCP_BUCKET",
	},
	Minio: &config.MinioConfig{
		Endpoint:  "127.0.0.1:9001",
		Key:       "minio",
		Secret:    "minio123",
		EnableSSL: false,
		Bucket:    "gomods",
	},
	Mongo: &config.MongoConfig{
		URL:                   "mongodb://127.0.0.1:27017",
		CertPath:              "",
		InsecureConn:          false,
		DefaultDBName:         "athens",
		DefaultCollectionName: "modules",
	},
	S3: &config.S3Config{
		Region: "MY_AWS_REGION",
		Key:    "MY_AWS_ACCESS_KEY_ID",
		Secret: "MY_AWS_SECRET_ACCESS_KEY",
		Token:  "",
		Bucket: "MY_S3_BUCKET_NAME",
	},
}

var testConfig = &config.Config{
	GoEnv:           "development",
	LogLevel:        "debug",
	GoBinary:        "go",
	GoProxy:         "direct",
	GoGetWorkers:    10,
	ProtocolWorkers: 30,
	CloudRuntime:    "none",
	TimeoutConf: config.TimeoutConf{
		Timeout: 300,
	},
	StorageType:      "memory",
	GlobalEndpoint:   "http://localhost:3001",
	Port:             ":3000",
	EnablePprof:      false,
	PprofPort:        ":3001",
	BasicAuthUser:    "",
	BasicAuthPass:    "",
	Storage:          testStorage,
	TraceExporterURL: "http://localhost:14268",
	TraceExporter:    "",
	StatsExporter:    "prometheus",
	SingleFlightType: "memory",
	GoBinaryEnvVars:  []string{"GOPROXY=direct"},
	SingleFlight:     &config.SingleFlight{},
	SumDBs:           []string{"https://sum.golang.org"},
	NoSumPatterns:    []string{},
	DownloadMode:     "sync",
	RobotsFile:       "robots.txt",
}

func TestConfigHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/config", nil)
	if err != nil {
		t.Fatal(err)
	}

	h := ConfigHandler(testConfig)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	toCheck := &config.Config{}
	if err := json.Unmarshal(rr.Body.Bytes(), toCheck); err != nil {
		t.Error("Failed to unmarshal the reply")
	}
	if !cmp.Equal(toCheck, testConfig) {
		diff := cmp.Diff(toCheck, testConfig)
		t.Errorf("Returned config different from original. Diff: %s", diff)
	}
}

func TestConfigHandlerWrongPath(t *testing.T) {
	req, err := http.NewRequest("GET", "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	h := ConfigHandler(testConfig)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusForbidden)
	}
}
