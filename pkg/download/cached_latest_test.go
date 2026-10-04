package download

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/module"
)

const cachedTestModule = "example.com/offline"

func TestDiscoveryFromPrefilledDisk(t *testing.T) {
	const version = "v1.0.0"
	root := t.TempDir()
	dir := filepath.Join(root, cachedTestModule, version)
	require.NoError(t, os.MkdirAll(dir, 0o700))
	info := &storage.RevInfo{Version: version, Time: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)}
	data, err := json.Marshal(info)
	require.NoError(t, err)
	modFile := []byte("module " + cachedTestModule + "\n\ngo 1.24\n")
	for name, contents := range map[string][]byte{
		version + ".info": data,
		"go.mod":          modFile,
		"source.zip":      cachedTestArchive(t, version, modFile),
	} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), contents, 0o600))
	}
	// Open storage only after placing the files, as in an air-gapped transfer.
	backend, err := fs.NewStorage(root, afero.NewOsFs())
	require.NoError(t, err)
	for _, tc := range []struct {
		name         string
		networkMode  string
		downloadMode mode.Mode
		wantCalls    int64
	}{
		{"offline", Offline, mode.None, 0},
		{"none", Strict, mode.None, 0},
		{"fallback outage", Fallback, mode.Sync, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lister := &cachedLatestLister{err: errors.E("upstream", "upstream unavailable", errors.KindGatewayTimeout)}
			dp := New(&Opts{
				Storage: backend, Lister: lister, NetworkMode: tc.networkMode,
				DownloadFile: &mode.DownloadFile{Mode: tc.downloadMode},
			})
			versions, err := dp.List(t.Context(), cachedTestModule)
			require.NoError(t, err)
			require.Equal(t, []string{version}, versions)
			latest, err := dp.Latest(t.Context(), cachedTestModule)
			require.NoError(t, err)
			require.Equal(t, info, latest)
			require.Equal(t, tc.wantCalls, lister.calls.Load())
		})
	}
}

type cachedLatestLister struct {
	info  *storage.RevInfo
	err   error
	calls atomic.Int64
}

func (l *cachedLatestLister) List(context.Context, string) (*storage.RevInfo, []string, error) {
	l.calls.Add(1)
	return l.info, nil, l.err
}

func saveCachedTestVersion(t *testing.T, backend storage.Backend, version string, modFile, zipFile []byte) *storage.RevInfo {
	t.Helper()
	stamp := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	if module.IsPseudoVersion(version) {
		var err error
		stamp, err = module.PseudoVersionTime(version)
		require.NoError(t, err)
	}
	info := &storage.RevInfo{Version: version, Time: stamp}
	data, err := json.Marshal(info)
	require.NoError(t, err)
	require.NoError(t, backend.Save(t.Context(), cachedTestModule, version, modFile, bytes.NewReader(zipFile), nil, data))
	return info
}

func TestCachedLatestSelection(t *testing.T) {
	for _, tc := range []struct {
		name     string
		versions []string
		want     string
	}{
		{"semantic release order", []string{"v1.2.0", "v1.10.0"}, "v1.10.0"},
		{"release before prerelease without commits", []string{"v1.2.0", "v1.3.0-rc1"}, "v1.2.0"},
		{"prereleases without releases", []string{"v1.2.0-rc1", "v1.2.0-rc2"}, "v1.2.0-rc2"},
		{"pseudo versions by time instead of base version", []string{
			"v1.10.1-0.20261001000000-abcdef123456",
			"v1.2.1-0.20261002000000-abcdef123456",
		}, "v1.2.1-0.20261002000000-abcdef123456"},
		{"commit fallback even with a newer release", []string{
			"v1.2.1-0.20261001000000-abcdef123456", "v1.3.0",
		}, "v1.2.1-0.20261001000000-abcdef123456"},
		{"preserve incompatible suffix", []string{"v1.0.0", "v2.0.0+incompatible"}, "v2.0.0+incompatible"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			backend, err := mem.NewStorage()
			require.NoError(t, err)
			var want *storage.RevInfo
			for _, version := range tc.versions {
				info := saveCachedTestVersion(t, backend, version, nil, nil)
				if version == tc.want {
					want = info
				}
			}
			lister := &cachedLatestLister{}
			dp := New(&Opts{Storage: backend, Lister: lister, NetworkMode: Offline})
			got, err := dp.Latest(t.Context(), cachedTestModule)
			require.NoError(t, err)
			require.Equal(t, want, got)
			require.Zero(t, lister.calls.Load(), "offline lookup must not query upstream")
		})
	}
}

func TestCachedLatestPolicy(t *testing.T) {
	upstreamErr := errors.E("upstream", "upstream unavailable", errors.KindGatewayTimeout)
	upstreamInfo := &storage.RevInfo{Version: "v1.1.0"}
	for _, tc := range []struct {
		name         string
		networkMode  string
		downloadMode mode.Mode
		cached       bool
		upstreamErr  error
		wantVersion  string
		wantKind     int
		wantCalls    int64
	}{
		{"offline hit", Offline, mode.Sync, true, nil, "v1.0.0", 0, 0},
		{"offline miss", Offline, mode.Sync, false, nil, "", errors.KindNotFound, 0},
		{"none with strict", Strict, mode.None, true, nil, "v1.0.0", 0, 0},
		{"none with fallback", Fallback, mode.None, true, nil, "v1.0.0", 0, 0},
		{"none miss", Strict, mode.None, false, nil, "", errors.KindNotFound, 0},
		{"fallback upstream succeeds", Fallback, mode.Sync, true, nil, "v1.1.0", 0, 1},
		{"fallback upstream fails", Fallback, mode.Sync, true, upstreamErr, "v1.0.0", 0, 1},
		{"fallback repository missing", Fallback, mode.Sync, true,
			errors.E("upstream", "remote: Repository not found", errors.KindNotFound), "v1.0.0", 0, 1},
		{"fallback miss preserves upstream error", Fallback, mode.Sync, false, upstreamErr, "", errors.KindGatewayTimeout, 1},
		{"strict error stays an error", Strict, mode.Sync, true, upstreamErr, "", errors.KindGatewayTimeout, 1},
		{"strict uses upstream", Strict, mode.Sync, true, nil, "v1.1.0", 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			backend, err := mem.NewStorage()
			require.NoError(t, err)
			if tc.cached {
				saveCachedTestVersion(t, backend, "v1.0.0", nil, nil)
			}
			lister := &cachedLatestLister{info: upstreamInfo, err: tc.upstreamErr}
			dp := New(&Opts{
				Storage: backend, Lister: lister, NetworkMode: tc.networkMode,
				DownloadFile: &mode.DownloadFile{Mode: tc.downloadMode},
			})
			got, err := dp.Latest(t.Context(), cachedTestModule)
			if tc.wantKind != 0 {
				require.Error(t, err)
				require.Equal(t, tc.wantKind, errors.Kind(err))
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantVersion, got.Version)
			}
			require.Equal(t, tc.wantCalls, lister.calls.Load())
		})
	}
}

func TestCachedLatestRejectsBadMetadata(t *testing.T) {
	for _, data := range []string{"not JSON", `{}`, `{"Version":"v1.1.0"}`} {
		t.Run(data, func(t *testing.T) {
			backend, err := mem.NewStorage()
			require.NoError(t, err)
			require.NoError(t, backend.Save(t.Context(), cachedTestModule, "v1.0.0", nil, bytes.NewReader(nil), nil, []byte(data)))
			lister := &cachedLatestLister{}
			dp := New(&Opts{Storage: backend, Lister: lister, NetworkMode: Offline})
			got, err := dp.Latest(t.Context(), cachedTestModule)
			require.Error(t, err)
			require.Equal(t, errors.KindUnexpected, errors.Kind(err))
			require.Nil(t, got)
			require.Zero(t, lister.calls.Load())
		})
	}
}

func TestCachedListOmitsPseudoVersions(t *testing.T) {
	for _, tc := range []struct {
		name         string
		networkMode  string
		downloadMode mode.Mode
		upstreamErr  error
	}{
		{"offline", Offline, mode.Sync, nil},
		{"none", Strict, mode.None, nil},
		{"fallback failure", Fallback, mode.Sync, errors.E("upstream", "unavailable")},
		{"fallback missing repository", Fallback, mode.Sync, errors.E("upstream", "remote: Repository not found", errors.KindNotFound)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			backend, err := mem.NewStorage()
			require.NoError(t, err)
			saveCachedTestVersion(t, backend, "v1.0.0", nil, nil)
			saveCachedTestVersion(t, backend, "v1.0.1-0.20261002000000-abcdef123456", nil, nil)
			dp := New(&Opts{
				Storage: backend, Lister: &cachedLatestLister{err: tc.upstreamErr},
				NetworkMode: tc.networkMode, DownloadFile: &mode.DownloadFile{Mode: tc.downloadMode},
			})
			got, err := dp.List(t.Context(), cachedTestModule)
			require.NoError(t, err)
			require.Equal(t, []string{"v1.0.0"}, got)
		})
	}
}
