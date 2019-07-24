package mongo

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
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

func TestQueryModuleVersionExists(t *testing.T) {
	modname, ver := "getTestModule", "v1.2.3"
	mock := &storage.Version{
		Info: []byte("123"),
		Mod:  []byte("456"),
		Zip:  ioutil.NopCloser(bytes.NewReader([]byte("789"))),
	}

	ctx := context.Background()
	backend := getStorage(t)

	zipBts, _ := ioutil.ReadAll(mock.Zip)
	backend.Save(ctx, modname, ver, mock.Mod, bytes.NewReader(zipBts), mock.Info)
	defer backend.Delete(ctx, modname, ver)

	info, err := query(ctx, backend, modname, ver)
	require.NoError(t, err)
	require.Equal(t, mock.Info, info.Info)
	require.Equal(t, mock.Mod, info.Mod)
}

func TestQueryKindNotFoundErrorCases(t *testing.T) {
	modname, ver := "getTestModule", "v1.2.3"
	mock := &storage.Version{
		Info: []byte("123"),
		Mod:  []byte("456"),
		Zip:  ioutil.NopCloser(bytes.NewReader([]byte("789"))),
	}

	ctx := context.Background()
	backend := getStorage(t)

	zipBts, _ := ioutil.ReadAll(mock.Zip)
	backend.Save(ctx, modname, ver, mock.Mod, bytes.NewReader(zipBts), mock.Info)
	defer backend.Delete(ctx, modname, ver)

	testCases := []struct {
		modname string
		ver     string
	}{
		{"getTestModule", "yyy"}, // test NotFound non-existent version
		{"getTestModule", ""},    // test NotFound empty str version
		{"xxx", "v1.2.3"},        // test NotFound non-existent module
		{"", "v1.2.3"},           // test NotFound empty str module
		{"", ""},                 // test NotFound empty str module and version
		{"xxx", "yyy"},           // test NotFound non-existent module and version
	}

	for _, test := range testCases {
		_, err := query(ctx, backend, test.modname, test.ver)
		require.Error(t, err)
		require.Equal(t, errors.KindNotFound, errors.Kind(err))
	}
}
