package godl

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/stretchr/testify/require"
)

const archive = "go1.25.1.linux-amd64.tar.gz"

func get(t *testing.T, h http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
	return w
}

type testUpstream struct {
	body                       []byte
	archiveCalls, listingCalls atomic.Int64
	status                     atomic.Int64
	listing                    []storage.Release
}

func (u *testUpstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if status := u.status.Load(); status != 0 {
		w.WriteHeader(int(status))
		return
	}
	if r.URL.Path == "/dl/" {
		u.listingCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(u.listing)
		return
	}
	u.archiveCalls.Add(1)
	if r.URL.Path != "/dl/"+archive {
		http.NotFound(w, r)
		return
	}
	_, _ = w.Write(u.body)
}

func handlerFixture(t *testing.T) (*Handler, *testUpstream, *httptest.Server) {
	t.Helper()
	body := []byte("verified archive bytes")
	up := &testUpstream{body: body, listing: []storage.Release{testRelease(body)}}
	server := httptest.NewServer(up)
	t.Cleanup(server.Close)
	u, err := url.Parse(server.URL + "/dl")
	require.NoError(t, err)
	backend, err := mem.NewStorage()
	require.NoError(t, err)
	h, err := New(Options{Storage: backend.(storage.ToolchainStorage), Upstream: u, Source: "official", Client: server.Client()})
	require.NoError(t, err)
	t.Cleanup(h.service.Close)
	return h, up, server
}

func TestListingAndArchivePersistence(t *testing.T) {
	h, up, _ := handlerFixture(t)
	for range 2 {
		w := get(t, h, "/?mode=json&include=all")
		require.Equal(t, 200, w.Code)
		require.Contains(t, w.Body.String(), "go1.25.1")
		require.Equal(t, "application/json", w.Header().Get("Content-Type"))
		w = get(t, h, "/"+archive)
		require.Equal(t, 200, w.Code)
		require.Equal(t, string(up.body), w.Body.String())
	}
	require.Equal(t, int64(1), up.archiveCalls.Load())
	require.Equal(t, int64(1), up.listingCalls.Load())
	opts := h.service.opts
	opts.NetworkMode = download.Offline
	opts.Upstream = nil
	restored, err := New(opts)
	require.NoError(t, err)
	require.Equal(t, 200, get(t, restored, "/"+archive).Code)
	require.Equal(t, int64(1), up.archiveCalls.Load())
}

func TestConcurrentMissesShareOneDownload(t *testing.T) {
	h, up, _ := handlerFixture(t)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() { require.Equal(t, 200, get(t, h, "/"+archive).Code) })
	}
	wg.Wait()
	require.Equal(t, int64(1), up.archiveCalls.Load())
}

func TestArchiveHTTP(t *testing.T) {
	h, up, _ := handlerFixture(t)
	first := get(t, h, "/"+archive)
	require.Equal(t, "private, no-cache", first.Header().Get("Cache-Control"))
	require.Equal(t, 200, first.Code)
	cases := []struct {
		method, header, value string
		code                  int
		body                  string
	}{
		{http.MethodHead, "", "", 200, ""}, {http.MethodGet, "Range", "bytes=2-5", 206, string(up.body[2:6])}, {http.MethodGet, "Range", "bytes=-3", 206, string(up.body[len(up.body)-3:])}, {http.MethodGet, "Range", "bytes=999-", 416, "invalid range: failed to overlap\n"}, {http.MethodGet, "If-None-Match", first.Header().Get("ETag"), 304, ""},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, "/"+archive, nil)
		if tc.header != "" {
			req.Header.Set(tc.header, tc.value)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		require.Equal(t, tc.code, w.Code)
		require.Equal(t, tc.body, w.Body.String())
	}
	require.Equal(t, int64(1), up.archiveCalls.Load())
}

func TestRejectedRequestsNeverFetch(t *testing.T) {
	h, up, _ := handlerFixture(t)
	for _, p := range []string{"/", "/?mode=text", "/foo", "/../etc/passwd", "/" + archive + "/../../x", "/GO1.25.1.linux-amd64.tar.gz", "/" + archive + ".sha256"} {
		require.Equal(t, 404, get(t, h, p).Code, p)
	}
	require.Equal(t, 400, get(t, h, "/?mode=json&include=all&include=x").Code)
	require.Equal(t, 400, get(t, h, "/?mode=json&unknown=x").Code)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/"+archive, nil))
	require.Equal(t, 405, w.Code)
	require.Zero(t, up.archiveCalls.Load())
	require.Zero(t, up.listingCalls.Load())
}

func TestMissingAndCorruptUpstream(t *testing.T) {
	h, up, _ := handlerFixture(t)
	require.Equal(t, 404, get(t, h, "/go9.9.9.linux-amd64.tar.gz").Code)
	require.Zero(t, up.archiveCalls.Load())
	// Valid metadata does not make corrupt bytes a committed storage hit.
	up.body = []byte("wrong bytes")
	require.Equal(t, 502, get(t, h, "/"+archive).Code)
	files, err := h.service.opts.Storage.ListArchives(t.Context(), "official")
	require.NoError(t, err)
	require.Empty(t, files)
	up.body = []byte("verified archive bytes")
	require.Equal(t, 200, get(t, h, "/"+archive).Code)
}

func TestHTMLTruncationAndRedirect(t *testing.T) {
	for _, kind := range []string{"html", "truncated", "redirect"} {
		t.Run(kind, func(t *testing.T) {
			h, up, _ := handlerFixture(t)
			final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch kind {
				case "html":
					w.Header().Set("Content-Type", "text/html")
					_, _ = io.WriteString(w, "<html>not an archive</html>")
				case "truncated":
					w.Header().Set("Content-Length", "100")
					_, _ = io.WriteString(w, "short")
				default:
					_, _ = w.Write(up.body)
				}
			}))
			defer final.Close()
			proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("mode") == "json" {
					_ = json.NewEncoder(w).Encode(up.listing)
					return
				}
				http.Redirect(w, r, final.URL, http.StatusFound)
			}))
			defer proxy.Close()
			u, err := url.Parse(proxy.URL)
			require.NoError(t, err)
			h.service.opts.Upstream = u
			w := get(t, h, "/"+archive)
			switch kind {
			case "html":
				require.Equal(t, 404, w.Code)
			case "truncated":
				require.Equal(t, 502, w.Code)
			default:
				require.Equal(t, 200, w.Code)
			}
		})
	}
}

func TestStrictAndFallbackRefresh(t *testing.T) {
	h, up, _ := handlerFixture(t)
	require.Equal(t, 200, get(t, h, "/"+archive).Code)
	opts := h.service.opts
	opts.Storage = staleListingStorage{opts.Storage}
	up.status.Store(503)
	strict, err := New(opts)
	require.NoError(t, err)
	require.Equal(t, 502, get(t, strict, "/?mode=json&include=all").Code)
	opts.NetworkMode = download.Fallback
	fallback, err := New(opts)
	require.NoError(t, err)
	w := get(t, fallback, "/?mode=json&include=all")
	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), archive)
}

func TestAllDownloadModes(t *testing.T) {
	for _, m := range []mode.Mode{mode.Sync, mode.None, mode.Async, mode.Redirect, mode.AsyncRedirect} {
		t.Run(string(m), func(t *testing.T) {
			h, up, _ := handlerFixture(t)
			h.service.opts.DownloadMode = m
			w := get(t, h, "/"+archive)
			switch m {
			case mode.Sync:
				require.Equal(t, 200, w.Code)
			case mode.None:
				require.Equal(t, 404, w.Code)
				require.Zero(t, up.listingCalls.Load())
			case mode.Redirect:
				require.Equal(t, 307, w.Code)
				require.True(t, strings.HasSuffix(w.Header().Get("Location"), "/dl/"+archive))
				require.Zero(t, up.archiveCalls.Load())
			case mode.Async:
				require.Equal(t, 404, w.Code)
			case mode.AsyncRedirect:
				require.Equal(t, 307, w.Code)
			}
			if m == mode.Async || m == mode.AsyncRedirect {
				require.Eventually(t, func() bool {
					_, body, err := h.service.opts.Storage.Archive(t.Context(), "official", archive)
					if err != nil {
						return false
					}
					_ = body.Close()
					return true
				}, time.Second, 5*time.Millisecond)
			}
		})
	}
}

func TestDefaultListingOrdering(t *testing.T) {
	releases := []storage.Release{{Version: "go1.9.2", Stable: true}, {Version: "go1.10.1", Stable: true}, {Version: "go1.10.2", Stable: true}, {Version: "go1.11rc1", Stable: false}, {Version: "go1.8.9", Stable: true}}
	selected := selectReleases(releases, false)
	require.Equal(t, []string{"go1.10.2", "go1.9.2"}, []string{selected[0].Version, selected[1].Version})
}

func TestNewRejectsBadConfiguration(t *testing.T) {
	backend, err := mem.NewStorage()
	require.NoError(t, err)
	u, _ := url.Parse("ftp://example.com/dl")
	_, err = New(Options{Storage: backend.(storage.ToolchainStorage), Source: "official", Upstream: u})
	require.Error(t, err)
	_, err = New(Options{})
	require.Error(t, err)
}

func TestDisabled(t *testing.T) {
	w := get(t, Disabled(), "/"+archive)
	require.Equal(t, 422, w.Code)
	require.Contains(t, w.Body.String(), "ATHENS_GO_DOWNLOAD_URL")
}

func TestNoListingCacheQueryCollision(t *testing.T) {
	h, up, _ := handlerFixture(t)
	older := testRelease(up.body)
	older.Version = "go1.24.1"
	older.Files[0].Version = older.Version
	older.Files[0].Filename = "go1.24.1.linux-amd64.tar.gz"
	oldest := older
	oldest.Version = "go1.23.1"
	oldest.Files = append([]storage.ReleaseFile(nil), older.Files...)
	oldest.Files[0].Version = oldest.Version
	oldest.Files[0].Filename = "go1.23.1.linux-amd64.tar.gz"
	up.listing = append(up.listing, older, oldest)
	latest := get(t, h, "/?mode=json")
	all := get(t, h, "/?mode=json&include=all")
	require.NotContains(t, latest.Body.String(), "go1.23.1")
	require.Contains(t, all.Body.String(), "go1.23.1")
	require.Equal(t, int64(1), up.listingCalls.Load())
}

func TestInstallerOnlyInventoryIsNotAdvertised(t *testing.T) {
	h, up, _ := handlerFixture(t)
	file := testRelease(up.body).Files[0]
	file.Filename = "go1.25.1.linux-amd64.pkg"
	file.Kind = "installer"
	require.NoError(t, h.service.opts.Storage.SaveArchive(t.Context(), "official", storage.ArchiveInfo{ReleaseFile: file, Stable: true}, bytes.NewReader(up.body)))
	h.service.opts.NetworkMode = download.Offline
	require.Equal(t, "[]", get(t, h, "/?mode=json&include=all").Body.String())
}

func TestConflictingManifestIsRejected(t *testing.T) {
	for _, field := range []string{"size", "platform", "kind", "stable"} {
		t.Run(field, func(t *testing.T) {
			h, up, _ := handlerFixture(t)
			other := testRelease(up.body)
			switch field {
			case "size":
				other.Files[0].Size++
			case "platform":
				other.Files[0].OS = "darwin"
			case "kind":
				other.Files[0].Kind = "source"
			case "stable":
				other.Stable = false
			}
			up.listing = append(up.listing, other)
			require.Equal(t, 502, get(t, h, "/?mode=json&include=all").Code)
			require.Zero(t, up.archiveCalls.Load())
		})
	}
}

func TestListingOmitsUnverifiableHistoricalFiles(t *testing.T) {
	h, up, _ := handlerFixture(t)
	historical := testRelease(up.body)
	historical.Version = "go1.4.2"
	historical.Files[0].Filename = "go1.4.2.darwin-amd64-osx10.6.tar.gz"
	historical.Files[0].Version = historical.Version
	historical.Files[0].SHA256 = ""
	historical.Files[0].Size = 0
	up.listing = append(up.listing, historical)
	w := get(t, h, "/?mode=json&include=all")
	require.Equal(t, 200, w.Code)
	require.NotContains(t, w.Body.String(), "go1.4.2")
	require.Contains(t, w.Body.String(), "go1.25.1")
	historical.Files[0].Size = 100
	up.listing = []storage.Release{testRelease(up.body), historical}
	w = get(t, h, "/?mode=json&include=all")
	require.Equal(t, 200, w.Code)
	require.NotContains(t, w.Body.String(), "go1.4.2")
}

func TestDiscoveryPlacesArchivesBeforeInstallers(t *testing.T) {
	h, up, _ := handlerFixture(t)
	installer := up.listing[0].Files[0]
	installer.Kind = "installer"
	installer.OS = "windows"
	installer.Filename = "go1.25.1.windows-amd64.msi"
	up.listing[0].Files = append([]storage.ReleaseFile{installer}, up.listing[0].Files...)
	response := get(t, h, "/?mode=json&include=all")
	require.Equal(t, http.StatusOK, response.Code)
	var releases []storage.Release
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &releases))
	require.Equal(t, "archive", releases[0].Files[0].Kind)
	require.Equal(t, "installer", releases[0].Files[1].Kind)
}
