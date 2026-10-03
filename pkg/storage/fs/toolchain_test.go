package fs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func archiveInfo(body []byte) storage.ArchiveInfo {
	sum := sha256.Sum256(body)
	return storage.ArchiveInfo{ReleaseFile: storage.ReleaseFile{Filename: "go1.25.1.linux-amd64.tar.gz", Version: "go1.25.1", OS: "linux", Arch: "amd64", Kind: "archive", Size: int64(len(body)), SHA256: hex.EncodeToString(sum[:])}, Stable: true}
}

func TestToolchainStorage(t *testing.T) {
	for _, memory := range []bool{false, true} {
		t.Run(map[bool]string{false: "disk", true: "memory"}[memory], func(t *testing.T) {
			filesystem := afero.NewOsFs()
			root := t.TempDir()
			if memory {
				filesystem = afero.NewMemMapFs()
				root = "/storage"
				require.NoError(t, filesystem.MkdirAll(root, 0o750))
			}
			backend, err := NewStorage(root, filesystem)
			require.NoError(t, err)
			store, ok := backend.(storage.ToolchainStorage)
			require.True(t, ok, "configured backend must expose release storage")
			ctx := t.Context()
			body := []byte("verified archive")
			info := archiveInfo(body)
			_, _, err = store.Archive(ctx, "official", info.Filename)
			require.True(t, errors.Is(err, errors.KindNotFound))
			require.Error(t, store.SaveArchive(ctx, "official", info, bytes.NewReader([]byte("wrong bytes"))))
			files, err := store.ListArchives(ctx, "official")
			require.NoError(t, err)
			require.Empty(t, files)
			require.NoError(t, store.SaveArchive(ctx, "official", info, bytes.NewReader(body)))
			require.NoError(t, store.SaveArchive(ctx, "official", info, bytes.NewReader(body)))
			other := []byte("other verified bytes")
			err = store.SaveArchive(ctx, "official", archiveInfo(other), bytes.NewReader(other))
			require.True(t, errors.Is(err, errors.KindAlreadyExists))
			require.NoError(t, store.SaveArchive(ctx, "other-distribution", archiveInfo(other), bytes.NewReader(other)))
			metadata, rc, err := store.Archive(ctx, "official", info.Filename)
			require.NoError(t, err)
			got, err := io.ReadAll(rc)
			require.NoError(t, err)
			require.NoError(t, rc.Close())
			require.Equal(t, body, got)
			require.Equal(t, info.Size, rc.Size())
			require.False(t, metadata.SavedAt.IsZero())
			newer := &storage.ReleaseList{FetchedAt: time.Now().UTC(), Releases: []storage.Release{{Version: info.Version, Stable: true, Files: []storage.ReleaseFile{info.ReleaseFile}}}}
			older := &storage.ReleaseList{FetchedAt: newer.FetchedAt.Add(-time.Hour)}
			require.NoError(t, store.SaveReleases(ctx, "official", newer))
			require.NoError(t, store.SaveReleases(ctx, "official", older))
			// Reconstruct the backend to prove metadata and archives are persisted.
			restarted, err := NewStorage(root, filesystem)
			require.NoError(t, err)
			restored := restarted.(storage.ToolchainStorage)
			list, err := restored.Releases(ctx, "official")
			require.NoError(t, err)
			require.Equal(t, newer, list)
			catalog, _, err := backend.(storage.Cataloger).Catalog(ctx, "", 100)
			require.NoError(t, err)
			require.Empty(t, catalog)
			versions, err := backend.List(ctx, "golang.org/toolchain")
			require.NoError(t, err)
			require.Empty(t, versions)
			require.NoError(t, restored.DeleteArchive(ctx, "official", info.Filename))
			files, err = restored.ListArchives(ctx, "official")
			require.NoError(t, err)
			require.Empty(t, files)
		})
	}
}

func TestConcurrentArchivePublication(t *testing.T) {
	root := t.TempDir()
	body := []byte("concurrent verified archive")
	info := archiveInfo(body)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			backend, err := NewStorage(root, afero.NewOsFs())
			require.NoError(t, err)
			require.NoError(t, backend.(storage.ToolchainStorage).SaveArchive(t.Context(), "official", info, bytes.NewReader(body)))
		})
	}
	wg.Wait()
}

func TestInterruptedDeleteCanBeReplaced(t *testing.T) {
	root := t.TempDir()
	backend, err := NewStorage(root, afero.NewOsFs())
	require.NoError(t, err)
	body := []byte("verified archive")
	info := archiveInfo(body)
	store := backend.(storage.ToolchainStorage)
	require.NoError(t, store.SaveArchive(t.Context(), "official", info, bytes.NewReader(body)))
	// Simulate a legacy deletion interrupted after removing data but before
	// removing its directory. A new verified download must be publishable.
	objects := backend.(*storageImpl).toolchains().Objects
	keys, err := objects.Keys(t.Context(), ".athens/toolchains/v1/")
	require.NoError(t, err)
	require.Len(t, keys, 1)
	require.NoError(t, afero.NewOsFs().Remove(filepath.Join(root, keys[0], "data")))
	require.NoError(t, store.SaveArchive(t.Context(), "official", info, bytes.NewReader(body)))
	require.NoError(t, store.DeleteArchive(t.Context(), "official", info.Filename))
	require.NoError(t, store.SaveArchive(t.Context(), "official", info, bytes.NewReader(body)))
}
