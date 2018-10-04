package fs

import (
	"os"
	"testing"

	"github.com/gomods/athens/pkg/storage/compliance"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestBackend(t *testing.T) {
	fs := afero.NewOsFs()
	b := getStorage(t, fs)
	compliance.RunTests(t, b, b.clear)
	fs.RemoveAll(b.rootDir)
}

func BenchmarkBackend(b *testing.B) {
	fs := afero.NewOsFs()
	backend := getStorage(b, fs)
	compliance.RunBenchmarks(b, backend, backend.clear)
	fs.RemoveAll(backend.rootDir)
}

func BenchmarkMemory(b *testing.B) {
	backend := getStorage(b, afero.NewMemMapFs())
	compliance.RunBenchmarks(b, backend, backend.clear)
}

func (s *storageImpl) clear() error {
	if err := s.filesystem.RemoveAll(s.rootDir); err != nil {
		return err
	}
	return s.filesystem.Mkdir(s.rootDir, os.ModeDir|os.ModePerm)
}

func getStorage(tb testing.TB, fs afero.Fs) *storageImpl {
	tb.Helper()
	dir, err := afero.TempDir(fs, "", "athens-fs-test")
	require.NoError(tb, err, "could not create temp dir")
	backend, err := NewStorage(dir, fs)
	require.NoError(tb, err)
	return backend.(*storageImpl)
}
