package mongo

import (
	"context"
	"os"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage/compliance"
	"github.com/stretchr/testify/require"
)

func TestBackend(t *testing.T) {
	backend := getStorage(t)
	compliance.RunTests(t, backend, backend.clear)
}

func (m *ModuleStore) clear() error {
	m.client.Database(m.db).Drop(context.Background())
	return nil
}

func BenchmarkBackend(b *testing.B) {
	backend := getStorage(b)
	compliance.RunBenchmarks(b, backend, backend.clear)
}

func getStorage(tb testing.TB) *ModuleStore {
	url := os.Getenv("ATHENS_MONGO_STORAGE_URL")
	if url == "" {
		tb.SkipNow()
	}
	backend, err := NewStorage(&config.MongoConfig{URL: url}, config.GetTimeoutDuration(300))
	require.NoError(tb, err)

	return backend
}

func TestMongoConfigVerification(t *testing.T) {
	testCases := []struct {
		testName     string
		url          string
		requireError bool
	}{
		{"Test Invalid Configuration:Empty URL", "", true},                         // test mongo configuration without url
		{"Test InValid Configuration:Misconfigured URL Scheme", "127.0.0.1", true}, // test mongo configuration with misconfigured url
		{"Test Valid Configuration: Full URL", "mongodb://127.0.0.1", false},       // test mongo configuration with url
	}

	for _, test := range testCases {
		t.Run(test.testName, func(t *testing.T) {
			backend := &ModuleStore{url: test.url, timeout: config.GetTimeoutDuration(300)}
			_, err := backend.newClient()
			if test.requireError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
