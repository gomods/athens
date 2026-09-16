package godl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const archive = "go1.27.1.linux-amd64.tar.gz"

// fakeUpstream mimics go.dev/dl: a JSON listing at /?mode=json and archives
// at /<name>, counting requests so tests can assert on cache behaviour.
type fakeUpstream struct {
	t        *testing.T
	files    map[string][]byte
	requests atomic.Int64
	listing  atomic.Int64
	// redirectTo, when set, makes archive requests 302 there first, like
	// go.dev/dl does to dl.google.com.
	redirectTo *httptest.Server
}

func (f *fakeUpstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.requests.Add(1)

	if r.URL.Path == "/dl/" {
		f.listing.Add(1)

		if r.URL.Query().Get("mode") != "json" {
			http.NotFound(w, r)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		require.NoError(f.t, json.NewEncoder(w).Encode([]map[string]string{{"version": "go1.27.1"}}))

		return
	}

	name := filepath.Base(r.URL.Path)

	body, ok := f.files[name]
	if !ok {
		http.NotFound(w, r)

		return
	}

	if f.redirectTo != nil {
		http.Redirect(w, r, f.redirectTo.URL+"/"+name, http.StatusFound)

		return
	}

	_, _ = w.Write(body)
}

func newTestHandler(t *testing.T, up *fakeUpstream) (*Handler, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(up)
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL + "/dl/")
	require.NoError(t, err)

	h, err := New(u, filepath.Join(t.TempDir(), "cache"), srv.Client(), 0)
	require.NoError(t, err)

	return h, srv
}

func get(t *testing.T, h http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))

	return w
}

func TestListingIsProxiedAndCached(t *testing.T) {
	up := &fakeUpstream{t: t}
	h, _ := newTestHandler(t, up)

	for range 3 {
		w := get(t, h, "/?mode=json&include=all")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "go1.27.1")
	}

	assert.Equal(t, int64(1), up.listing.Load(), "listing should be fetched once within its TTL")
}

func TestArchiveIsCachedOnDisk(t *testing.T) {
	up := &fakeUpstream{t: t, files: map[string][]byte{archive: []byte("tarball bytes")}}
	h, _ := newTestHandler(t, up)

	for range 2 {
		w := get(t, h, "/"+archive)
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "tarball bytes", w.Body.String())
	}

	assert.Equal(t, int64(1), up.requests.Load(), "second request must be a cache hit")

	cached, err := os.ReadFile(filepath.Join(h.cacheDir, archive))
	require.NoError(t, err)
	assert.Equal(t, "tarball bytes", string(cached))

	entries, err := os.ReadDir(h.cacheDir)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "no temp files may be left behind")
}

func TestArchiveFollowsUpstreamRedirect(t *testing.T) {
	final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "from cdn")
	}))
	t.Cleanup(final.Close)

	up := &fakeUpstream{t: t, files: map[string][]byte{archive: nil}, redirectTo: final}
	h, _ := newTestHandler(t, up)

	w := get(t, h, "/"+archive)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "from cdn", w.Body.String())
}

func TestConcurrentMissesShareOneDownload(t *testing.T) {
	up := &fakeUpstream{t: t, files: map[string][]byte{archive: []byte("x")}}
	h, _ := newTestHandler(t, up)

	var wg sync.WaitGroup

	const n = 8

	codes := make([]int, n)

	for i := range n {
		wg.Add(1)

		go func() {
			defer wg.Done()

			codes[i] = get(t, h, "/"+archive).Code
		}()
	}

	wg.Wait()

	for _, code := range codes {
		assert.Equal(t, http.StatusOK, code)
	}

	assert.Equal(t, int64(1), up.requests.Load())
}

func TestUpstream404IsPassedThroughAndNotCached(t *testing.T) {
	up := &fakeUpstream{t: t, files: map[string][]byte{}}
	h, _ := newTestHandler(t, up)

	w := get(t, h, "/go9.9.9.linux-amd64.tar.gz")
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Publish it and make sure the miss was not remembered.
	up.files["go9.9.9.linux-amd64.tar.gz"] = []byte("now exists")
	w = get(t, h, "/go9.9.9.linux-amd64.tar.gz")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpstreamErrorIsBadGateway(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	h, err := New(u, t.TempDir(), srv.Client(), 0)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadGateway, get(t, h, "/"+archive).Code)
	assert.Equal(t, http.StatusInternalServerError, get(t, h, "/?mode=json").Code)
}

func TestTruncatedDownloadIsNotCached(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		_, _ = io.WriteString(w, "short")
	}))
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	dir := t.TempDir()
	h, err := New(u, dir, srv.Client(), 0)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadGateway, get(t, h, "/"+archive).Code)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestRejectsEverythingElse(t *testing.T) {
	up := &fakeUpstream{t: t, files: map[string][]byte{archive: []byte("x"), "etc/passwd": []byte("no")}}
	h, _ := newTestHandler(t, up)

	for _, target := range []string{
		"/",                                    // listing without mode=json
		"/?mode=text",                          // wrong mode
		"/foo",                                 // not an archive name
		"/../etc/passwd",                       // traversal attempt
		"/go1.27.1.linux-amd64.tar.gz/../../x", // traversal with a valid prefix
		"/GO1.27.1.linux-amd64.tar.gz",         // case matters, it is what go.dev publishes
		"/go1.27.1.linux-amd64.tar.gz.sha256",  // go.dev/dl serves HTML for these, not the checksum
	} {
		assert.Equal(t, http.StatusNotFound, get(t, h, target).Code, target)
	}

	assert.Equal(t, int64(0), up.requests.Load(), "rejected paths must never reach the upstream")

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/"+archive, nil))
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHTMLFromUpstreamIsNotAnArchive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, "<html>Redirecting to documentation...</html>")
	}))
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	dir := t.TempDir()
	h, err := New(u, dir, srv.Client(), 0)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, get(t, h, "/"+archive).Code)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestHeadArchive(t *testing.T) {
	up := &fakeUpstream{t: t, files: map[string][]byte{archive: []byte("abc\n")}}
	h, _ := newTestHandler(t, up)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodHead, "/"+archive, nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "4", w.Header().Get("Content-Length"))
	assert.Empty(t, w.Body.String())
}

func TestNewRejectsBadUpstream(t *testing.T) {
	u, err := url.Parse("ftp://example.com/dl")
	require.NoError(t, err)

	_, err = New(u, t.TempDir(), nil, 0)
	require.Error(t, err)
}

func TestListingHeaders(t *testing.T) {
	up := &fakeUpstream{t: t}
	h, _ := newTestHandler(t, up)

	w := get(t, h, "/?mode=json&include=all")
	require.Equal(t, http.StatusOK, w.Code)

	lastModified, err := http.ParseTime(w.Header().Get("Last-Modified"))
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), lastModified, 5*time.Second)

	var maxAge int
	_, err = fmt.Sscanf(w.Header().Get("Cache-Control"), "public, max-age=%d", &maxAge)
	require.NoError(t, err)
	assert.InDelta(t, DefaultListingTTL.Seconds(), maxAge, 5)
}

func TestStaleListingIsServedWhenRefreshFails(t *testing.T) {
	var fail atomic.Bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		_, _ = io.WriteString(w, `[{"version":"go1.27.1"}]`)
	}))
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	// A one-nanosecond TTL makes every request a refresh attempt.
	h, err := New(u, t.TempDir(), srv.Client(), time.Nanosecond)
	require.NoError(t, err)

	first := get(t, h, "/?mode=json")
	require.Equal(t, http.StatusOK, first.Code)

	fail.Store(true)

	stale := get(t, h, "/?mode=json")
	assert.Equal(t, http.StatusOK, stale.Code)
	assert.Equal(t, first.Body.String(), stale.Body.String())
	assert.Equal(t, first.Header().Get("Last-Modified"), stale.Header().Get("Last-Modified"))
	assert.Equal(t, "public, max-age=0", stale.Header().Get("Cache-Control"), "an expired listing must not be cached downstream")

	// A connection error, not only an HTTP error, must also fall back.
	srv.Close()
	assert.Equal(t, http.StatusOK, get(t, h, "/?mode=json").Code)
}

func TestArchiveHeaders(t *testing.T) {
	up := &fakeUpstream{t: t, files: map[string][]byte{archive: []byte("x")}}
	h, _ := newTestHandler(t, up)

	w := get(t, h, "/"+archive)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "public, max-age=31536000, immutable", w.Header().Get("Cache-Control"))

	_, err := http.ParseTime(w.Header().Get("Last-Modified"))
	require.NoError(t, err)
}

func TestDisabled(t *testing.T) {
	w := get(t, Disabled(), "/"+archive)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "ATHENS_GO_DOWNLOAD_URL")
}
