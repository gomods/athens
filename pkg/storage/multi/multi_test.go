package multi

import (
	"os"
	"testing"

	"github.com/gomods/athens/pkg/storage"

	"github.com/gomods/athens/pkg/storage/compliance"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestBackend(t *testing.T) {
	filesystem := afero.NewMemMapFs()
	dir, err := afero.TempDir(filesystem, "", "athens-multi-test")
	require.NoError(t, err, "could not create temp dir")

	b := getStorage(t, filesystem, dir)
	compliance.RunTests(t, b, clear(filesystem, dir))
	filesystem.RemoveAll(dir)
}

func BenchmarkBackend(b *testing.B) {
	filesystem := afero.NewOsFs()
	dir, err := afero.TempDir(filesystem, "", "athens-multi-test")
	require.NoError(b, err, "could not create temp dir")

	backend := getStorage(b, filesystem, dir)
	compliance.RunBenchmarks(b, backend, clear(filesystem, dir))
	filesystem.RemoveAll(dir)
}

func BenchmarkMemory(b *testing.B) {
	filesystem := afero.NewMemMapFs()
	dir, err := afero.TempDir(filesystem, "", "athens-multi-test")
	require.NoError(b, err, "could not create temp dir")

	backend := getStorage(b, filesystem, dir)
	compliance.RunBenchmarks(b, backend, clear(filesystem, dir))
}

func clear(fs afero.Fs, rootDir string) func() error {
	return func() error {
		if err := fs.RemoveAll(rootDir); err != nil {
			return err
		}
		return fs.Mkdir(rootDir, os.ModeDir|os.ModePerm)
	}
}

func getStorage(tb testing.TB, filesystem afero.Fs, dir string) *Storage {
	tb.Helper()
	fsStore, err := fs.NewStorage(dir, filesystem)
	require.NoError(tb, err)

	backend, err := NewStorage([]storage.Backend{fsStore})
	require.NoError(tb, err)

	return backend
}
