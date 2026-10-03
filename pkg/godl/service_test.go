package godl

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/stretchr/testify/require"
)

func testRelease(body []byte) storage.Release {
	sum := sha256.Sum256(body)
	return storage.Release{Version: "go1.25.1", Stable: true, Files: []storage.ReleaseFile{{Filename: "go1.25.1.linux-amd64.tar.gz", OS: "linux", Arch: "amd64", Version: "go1.25.1", SHA256: hex.EncodeToString(sum[:]), Size: int64(len(body)), Kind: "archive"}}}
}

func TestServicePolicyAndRestart(t *testing.T) {
	for _, network := range []string{download.Strict, download.Fallback, download.Offline} {
		for _, m := range []mode.Mode{mode.Sync, mode.None} {
			t.Run(network+"/"+string(m), func(t *testing.T) {
				body := []byte("verified release archive")
				release := testRelease(body)
				var calls atomic.Int64
				var failed atomic.Bool
				upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					if failed.Load() {
						w.WriteHeader(503)
						return
					}
					if r.URL.Query().Get("mode") == "json" {
						require.Equal(t, "all", r.URL.Query().Get("include"))
						require.NoError(t, json.NewEncoder(w).Encode([]storage.Release{release}))
						return
					}
					_, _ = w.Write(body)
				}))
				defer upstream.Close()
				u, err := url.Parse(upstream.URL)
				require.NoError(t, err)
				backend, err := mem.NewStorage()
				require.NoError(t, err)
				store := backend.(storage.ToolchainStorage)
				opts := Options{Storage: store, Upstream: u, Source: "official", NetworkMode: network, DownloadMode: m, ListingTTL: time.Nanosecond}
				service, err := NewService(opts)
				require.NoError(t, err)
				t.Cleanup(service.Close)
				_, downloaded, err := service.Archive(t.Context(), release.Files[0].Filename)
				if network == download.Offline || m == mode.None {
					require.True(t, errors.Is(err, errors.KindNotFound))
					require.Zero(t, calls.Load())
				} else {
					require.NoError(t, err)
					require.NoError(t, downloaded.Close())
				}
				require.NoError(t, store.SaveArchive(t.Context(), "official", storage.ArchiveInfo{ReleaseFile: release.Files[0], Stable: true}, bytes.NewReader(body)))
				opts.NetworkMode = download.Offline
				opts.Upstream = nil
				restored, err := NewService(opts)
				require.NoError(t, err)
				t.Cleanup(restored.Close)
				before := calls.Load()
				listing, _, err := restored.Listing(t.Context(), true)
				require.NoError(t, err)
				require.Equal(t, []storage.Release{release}, listing)
				_, rc, err := restored.Archive(t.Context(), release.Files[0].Filename)
				require.NoError(t, err)
				got, err := io.ReadAll(rc)
				require.NoError(t, err)
				require.NoError(t, rc.Close())
				require.Equal(t, body, got)
				require.Equal(t, before, calls.Load())
				// A newer upstream version must not appear in storage-only discovery.
				newer := testRelease(body)
				newer.Version = "go1.25.2"
				newer.Files[0].Version = newer.Version
				newer.Files[0].Filename = "go1.25.2.linux-amd64.tar.gz"
				require.NoError(t, store.SaveReleases(t.Context(), "official", &storage.ReleaseList{FetchedAt: time.Now(), Releases: []storage.Release{newer, release}}))
				listing, _, err = restored.Listing(t.Context(), true)
				require.NoError(t, err)
				require.Equal(t, []storage.Release{release}, listing)
				failed.Store(true)
				if network != download.Offline && m != mode.None {
					online, err := NewService(optsWithNetwork(opts, network, u))
					require.NoError(t, err)
					t.Cleanup(online.Close)
					_, _, err = online.Listing(t.Context(), true)
					if network == download.Strict {
						require.Error(t, err)
					} else {
						require.NoError(t, err)
					}
				}
			})
		}
	}
}

func optsWithNetwork(opts Options, network string, u *url.URL) Options {
	opts.NetworkMode = network
	opts.Upstream = u
	return opts
}

func TestBackgroundFillStopsOnServiceShutdown(t *testing.T) {
	body := []byte("verified release archive")
	started := make(chan struct{})
	stopped := make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("mode") == "json" {
			_ = json.NewEncoder(w).Encode([]storage.Release{testRelease(body)})
			return
		}
		close(started)
		<-r.Context().Done()
		close(stopped)
	}))
	defer upstream.Close()
	u, err := url.Parse(upstream.URL)
	require.NoError(t, err)
	backend, err := mem.NewStorage()
	require.NoError(t, err)
	store := backend.(storage.ToolchainStorage)
	service, err := NewService(Options{Storage: store, Source: "official", Upstream: u, DownloadMode: mode.Async})
	require.NoError(t, err)
	defer service.Close()
	_, _, err = service.Archive(t.Context(), archive)
	require.True(t, errors.Is(err, errors.KindNotFound))
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("background download did not start")
	}
	service.Close()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown did not cancel upstream download")
	}
	files, err := store.ListArchives(t.Context(), "official")
	require.NoError(t, err)
	require.Empty(t, files)
}
