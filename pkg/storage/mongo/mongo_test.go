package mongo

import (
	"path/filepath"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage/compliance"
	"github.com/stretchr/testify/require"
)

var (
	testConfigFile = filepath.Join("..", "..", "..", "config.dev.toml")
)

func TestBackend(t *testing.T) {
	backend := getStorage(t)
	compliance.RunTests(t, backend, backend.clear)
}

func (m *ModuleStore) clear() error {
	m.s.Database(m.d).Drop()
	return m.initDatabase()
}

func BenchmarkBackend(b *testing.B) {
	backend := getStorage(b)
	compliance.RunBenchmarks(b, backend, backend.clear)
}

func getStorage(tb testing.TB) *ModuleStore {
	conf, err := config.GetConf(testConfigFile)
	if err != nil {
		tb.Fatalf("Unable to parse config file: %s", err.Error())
	}
	if conf.Storage.Mongo.URL == "" {
		tb.SkipNow()
	}

	backend, err := NewStorage(&config.MongoConfig{URL: conf.Storage.Mongo.URL}, config.GetTimeoutDuration(300))
	require.NoError(tb, err)

	return backend
}
