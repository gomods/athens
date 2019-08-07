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

func TestNewStorageWithNonDefaultDBName(t *testing.T) {
	url := os.Getenv("ATHENS_MONGO_STORAGE_URL")

	if url == "" {
		t.SkipNow()
	}

	testCases := []struct {
		name      string
		dbName    string
		expDbName string
	}{
		{"Test Default 'Athens' DB Name", "athens", "athens"}, //Tests the default database name
		{"Test Custom DB Name", "testAthens", "testAthens"},   //Tests a non-default database name
		{"Test Blank DB Name", "", "athens"},                  //Tests the blank database name edge-case
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			backend, err := NewStorage(&config.MongoConfig{URL: url, DefaultDBName: test.dbName}, config.GetTimeoutDuration(300))
			require.NoError(t, err)
			require.Equal(t, test.expDbName, backend.db)
		})
	}
}
