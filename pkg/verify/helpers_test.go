package verify_test

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"os"
	"testing"

	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/gomods/athens/pkg/verify"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/sumdb/dirhash"
)

// fakeOracle returns canonical hashes from a static map; unknown keys are
// reported as ErrNotInSumDB.
type fakeOracle struct{ hashes map[string]string }

func (o fakeOracle) CanonicalHash(_ context.Context, mod, ver string) (string, error) {
	if h, ok := o.hashes[mod+"@"+ver]; ok {
		return h, nil
	}
	return "", verify.ErrNotInSumDB
}

// writeZip builds an in-memory zip from name->body entries.
func writeZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range files {
		w, err := zw.Create(name)
		require.NoError(t, err)
		_, err = w.Write([]byte(body))
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

// hashOf returns the dirhash h1: of a zip's bytes.
func hashOf(t *testing.T, zipBytes []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.zip")
	require.NoError(t, err)
	_, err = f.Write(zipBytes)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	h, err := dirhash.HashZip(f.Name(), dirhash.Hash1)
	require.NoError(t, err)
	return h
}

// newTestStore returns an in-memory fs-backed storage backend (implements
// Cataloger/Getter/Saver/Deleter).
func newTestStore(t *testing.T) storage.Backend {
	t.Helper()
	memfs := afero.NewMemMapFs()
	err := memfs.Mkdir("/athens", 0755)
	require.NoError(t, err)
	b, err := fs.NewStorage("/athens", memfs)
	require.NoError(t, err)
	return b
}

// saveModule seeds a module version into the backend.
func saveModule(t *testing.T, ctx context.Context, b storage.Backend, mod, ver string, zipBytes []byte) {
	t.Helper()
	sum := md5.Sum(zipBytes)
	info := []byte(`{"Version":"` + ver + `","Time":"2025-01-01T00:00:00Z"}`)
	err := b.Save(ctx, mod, ver, []byte("module "+mod+"\n"), bytes.NewReader(zipBytes), sum[:], info)
	require.NoError(t, err)
}
