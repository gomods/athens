package compliance

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/gomods/athens/pkg/storage"
	"github.com/stretchr/testify/require"
)

// RunBenchmarks takes a backend and runs benchmarks against
// saving and loading modules.
func RunBenchmarks(b *testing.B, s storage.Backend, reset func() error) {
	b.Helper()
	benchList(b, s, reset)
	benchSave(b, s, reset)
	benchDelete(b, s, reset)
	benchExists(b, s, reset)
}

func benchList(b *testing.B, s storage.Backend, reset func() error) {
	b.Helper()
	require.NoError(b, reset())
	defer require.NoError(b, reset())
	module, version := "benchListModule", "1.0.1"
	mock := getMockModule()
	err := s.Save(
		context.Background(),
		module,
		version,
		mock.Mod,
		mock.Zip,
		mock.Info,
	)
	require.NoError(b, err, "save for storage failed")

	b.Run("list", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := s.List(context.Background(), module)
			require.NoError(b, err, "Error in listing module")
		}
	})
}

func benchSave(b *testing.B, s storage.Backend, reset func() error) {
	b.Helper()
	require.NoError(b, reset())
	defer require.NoError(b, reset())

	module, version := "benchSaveModule", "1.0.1"
	mock := getMockModule()
	zipBts, err := io.ReadAll(mock.Zip)
	require.NoError(b, err)

	mi := 0
	ctx := context.Background()
	b.Run("save", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := s.Save(
				ctx,
				fmt.Sprintf("save-%s-%d", module, mi),
				version,
				mock.Mod,
				bytes.NewReader(zipBts),
				mock.Info,
			)
			require.NoError(b, err)
			mi++
		}
	})
}

func benchDelete(b *testing.B, s storage.Backend, reset func() error) {
	b.Helper()
	require.NoError(b, reset())
	defer require.NoError(b, reset())

	module, version := "benchDeleteModule", "1.0.1"
	mock := getMockModule()
	zipBts, err := io.ReadAll(mock.Zip)
	require.NoError(b, err)
	ctx := context.Background()

	b.Run("delete", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			name := fmt.Sprintf("del-%s-%d", module, i)
			err := s.Save(ctx, name, version, mock.Mod, bytes.NewReader(zipBts), mock.Info)
			require.NoError(b, err, "saving %s for storage failed", name)
			err = s.Delete(ctx, name, version)
			require.NoError(b, err, "delete failed: %s", name)
		}
	})
}

func benchExists(b *testing.B, s storage.Backend, reset func() error) {
	b.Helper()
	require.NoError(b, reset())
	defer require.NoError(b, reset())

	module, version := "benchExistsModule", "1.0.1"
	mock := getMockModule()

	ctx := context.Background()
	err := s.Save(ctx, module, version, mock.Mod, mock.Zip, mock.Info)
	require.NoError(b, err)

	b.Run("exists", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			exists, err := storage.WithChecker(s).Exists(ctx, module, version)
			require.NoError(b, err)
			require.True(b, exists)
		}
	})
}
