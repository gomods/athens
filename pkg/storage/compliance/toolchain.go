package compliance

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// RunToolchainTests checks optional release storage using the same backend
// acceptance seam as module storage. Callers supply isolated test storage.
func RunToolchainTests(t *testing.T, backend storage.Backend) {
	t.Helper()
	store, ok := backend.(storage.ToolchainStorage)
	require.True(t, ok)
	source := "compliance-" + uuid.NewString()
	ctx := t.Context()
	body := []byte("verified archive contents")
	sum := sha256.Sum256(body)
	info := storage.ArchiveInfo{ReleaseFile: storage.ReleaseFile{Filename: "go1.25.1.linux-amd64.tar.gz", Version: "go1.25.1", OS: "linux", Arch: "amd64", Size: int64(len(body)), SHA256: hex.EncodeToString(sum[:]), Kind: "archive"}, Stable: true}
	_, _, err := store.Archive(ctx, source, info.Filename)
	require.True(t, errors.Is(err, errors.KindNotFound))
	require.Error(t, store.SaveArchive(ctx, source, info, bytes.NewReader([]byte("wrong"))))
	files, err := store.ListArchives(ctx, source)
	require.NoError(t, err)
	require.Empty(t, files)
	var wg sync.WaitGroup
	saveErrors := make(chan error, 4)
	for range 4 {
		wg.Go(func() { saveErrors <- store.SaveArchive(ctx, source, info, bytes.NewReader(body)) })
	}
	wg.Wait()
	close(saveErrors)
	for err := range saveErrors {
		require.NoError(t, err)
	}
	saved, rc, err := store.Archive(ctx, source, info.Filename)
	require.NoError(t, err)
	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.NoError(t, rc.Close())
	require.Equal(t, body, got)
	require.Equal(t, info.Size, rc.Size())
	require.False(t, saved.SavedAt.IsZero())
	conflict := info
	other := []byte("other bytes")
	otherSum := sha256.Sum256(other)
	conflict.Size = int64(len(other))
	conflict.SHA256 = hex.EncodeToString(otherSum[:])
	err = store.SaveArchive(ctx, source, conflict, bytes.NewReader(other))
	require.True(t, errors.Is(err, errors.KindAlreadyExists))
	newer := &storage.ReleaseList{FetchedAt: time.Now().UTC(), Releases: []storage.Release{{Version: info.Version, Stable: true, Files: []storage.ReleaseFile{info.ReleaseFile}}}}
	require.NoError(t, store.SaveReleases(ctx, source, newer))
	require.NoError(t, store.SaveReleases(ctx, source, &storage.ReleaseList{FetchedAt: newer.FetchedAt.Add(-time.Hour)}))
	listing, err := store.Releases(ctx, source)
	require.NoError(t, err)
	require.Equal(t, newer, listing)
	files, err = store.ListArchives(ctx, source)
	require.NoError(t, err)
	require.Len(t, files, 1)
	if catalog, ok := backend.(storage.Cataloger); ok {
		entries, _, err := catalog.Catalog(ctx, "", 1000)
		require.NoError(t, err)
		for _, entry := range entries {
			require.NotContains(t, entry.Module, ".athens/toolchains")
		}
	}
	require.NoError(t, store.DeleteArchive(ctx, source, info.Filename))
	files, err = store.ListArchives(ctx, source)
	require.NoError(t, err)
	require.Empty(t, files)
}
