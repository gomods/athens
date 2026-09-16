// Package godl proxies and caches Go toolchain release downloads (the
// contents of https://go.dev/dl) so that tools such as actions/setup-go can
// point their download base URL at Athens instead of the public internet.
package godl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// DefaultListingTTL is how long a fetched version listing is reused when
// New is given a zero TTL. New Go releases show up within this window for
// callers that resolve a version spec such as "1.25" to the newest patch.
const DefaultListingTTL = 2 * time.Hour

// archiveMaxAge is the Cache-Control max-age for archives: releases never
// change once published, so clients and intermediaries may keep them for as
// long as HTTP allows.
const archiveMaxAge = 365 * 24 * time.Hour

// archiveRE accepts the release archive names published under go.dev/dl,
// e.g. go1.27.1.linux-amd64.tar.gz or go1.27rc3.darwin-arm64.pkg. Anything
// else is rejected before touching the upstream or the cache directory.
// Checksum (.sha256) and signature (.asc) files are deliberately excluded:
// go.dev/dl answers those with a 200 HTML page rather than the file.
var archiveRE = regexp.MustCompile(`^go[0-9][A-Za-z0-9.\-]*\.(tar\.gz|zip|msi|pkg)$`)

// Handler serves /?mode=json version listings and release archives from an
// upstream go.dev/dl-compatible server, caching archives on local disk.
type Handler struct {
	upstream   *url.URL
	cacheDir   string
	client     *http.Client
	listingTTL time.Duration
	group      singleflight.Group

	listingMu      sync.Mutex
	listingBody    []byte
	listingType    string
	listingFetched time.Time
}

// New returns a Handler that fetches from upstream (e.g. https://go.dev/dl)
// and caches archives under cacheDir, creating it if needed. The cache is
// content-addressed by file name only; releases are immutable, so nothing is
// ever evicted or revalidated. listingTTL is how long the version listing is
// reused before being refetched; zero selects DefaultListingTTL.
func New(upstream *url.URL, cacheDir string, client *http.Client, listingTTL time.Duration) (*Handler, error) {
	if upstream.Scheme != "http" && upstream.Scheme != "https" {
		return nil, fmt.Errorf("godl: upstream %q must have an http or https scheme", upstream)
	}

	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		return nil, fmt.Errorf("godl: creating cache dir: %w", err)
	}

	if client == nil {
		client = http.DefaultClient
	}

	if listingTTL <= 0 {
		listingTTL = DefaultListingTTL
	}

	base := *upstream
	base.Path = strings.TrimSuffix(base.Path, "/")

	return &Handler{
		upstream:   &base,
		cacheDir:   cacheDir,
		client:     client,
		listingTTL: listingTTL,
	}, nil
}

// Disabled returns the handler Athens mounts at the same path when
// GoDownloadURL is unset, so clients get an explanation rather than a 404
// that looks like a missing Go version.
func Disabled() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w,
			"Go toolchain downloads are disabled on this Athens server: set GoDownloadURL (ATHENS_GO_DOWNLOAD_URL) to enable them",
			http.StatusUnprocessableEntity)
	})
}

// ServeHTTP implements http.Handler. It expects the path prefix Athens mounts
// it under to have been stripped already.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/")

	switch {
	case name == "" && r.URL.Query().Get("mode") == "json":
		h.serveListing(w, r)
	case name != "" && archiveRE.MatchString(name):
		h.serveArchive(w, r, name)
	default:
		http.NotFound(w, r)
	}
}

// serveListing proxies the version listing, keeping one copy in memory for
// listingTTL. The listing is small and changes rarely; the point is to keep
// every CI job from hitting the upstream for it. If refreshing an expired
// copy fails, the stale copy is served: an old listing is harmless compared
// to failing the request, and the next request retries the upstream.
func (h *Handler) serveListing(w http.ResponseWriter, r *http.Request) {
	h.listingMu.Lock()
	defer h.listingMu.Unlock()

	now := time.Now()

	if h.listingBody == nil || now.Sub(h.listingFetched) > h.listingTTL {
		body, contentType, status, err := h.fetchListing(r.Context(), r.URL.RawQuery)

		switch {
		case err == nil && status == http.StatusOK:
			h.listingBody = body
			h.listingType = contentType
			h.listingFetched = now
		case h.listingBody != nil:
			// Keep serving the stale copy below.
		case err != nil:
			http.Error(w, err.Error(), http.StatusBadGateway)

			return
		default:
			// Some upstreams (e.g. the Microsoft build of Go) publish no
			// listing; pass that through so clients fall back to direct
			// archive URLs instead of treating the mirror as broken.
			http.Error(w, http.StatusText(status), status)

			return
		}
	}

	// Let clients and intermediaries reuse the response until the next
	// refresh is due, and tell them how old what they got actually is.
	maxAge := max(h.listingFetched.Add(h.listingTTL).Sub(now), 0)

	w.Header().Set("Content-Type", h.listingType)
	w.Header().Set("Content-Length", fmt.Sprint(len(h.listingBody)))
	w.Header().Set("Last-Modified", h.listingFetched.UTC().Format(http.TimeFormat))
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds())))
	w.WriteHeader(http.StatusOK)

	if r.Method != http.MethodHead {
		_, _ = w.Write(h.listingBody)
	}
}

func (h *Handler) fetchListing(ctx context.Context, rawQuery string) ([]byte, string, int, error) {
	u := *h.upstream
	u.Path += "/"
	u.RawQuery = rawQuery

	resp, err := h.get(ctx, u.String())
	if err != nil {
		return nil, "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", resp.StatusCode, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", 0, fmt.Errorf("godl: reading listing from %s: %w", u.String(), err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}

	return body, contentType, http.StatusOK, nil
}

// serveArchive serves name from the cache directory, filling it from the
// upstream on a miss. Concurrent misses for the same file share one download.
func (h *Handler) serveArchive(w http.ResponseWriter, r *http.Request, name string) {
	local := filepath.Join(h.cacheDir, name)

	if _, err := os.Stat(local); err != nil {
		// The download is shared with every request waiting on this file,
		// so it must outlive the first requester's context.
		ctx := context.WithoutCancel(r.Context())

		_, err, _ = h.group.Do(name, func() (any, error) {
			return nil, h.fill(ctx, name, local)
		})

		var upstreamErr *upstreamStatusError
		switch {
		case errors.As(err, &upstreamErr):
			http.Error(w, err.Error(), upstreamErr.status)

			return
		case err != nil:
			http.Error(w, err.Error(), http.StatusBadGateway)

			return
		}
	}

	// ServeFile handles HEAD, Range and conditional requests, sets
	// Last-Modified from the file, and derives Content-Type from the extension.
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, immutable", int(archiveMaxAge.Seconds())))
	http.ServeFile(w, r, local)
}

// upstreamStatusError reports a non-200 upstream response for an archive. A
// 404 is passed through unchanged so clients see "no such version" rather
// than a proxy failure.
type upstreamStatusError struct {
	url    string
	status int
}

func (e *upstreamStatusError) Error() string {
	return fmt.Sprintf("godl: upstream %s returned %d", e.url, e.status)
}

// fill downloads name from the upstream into local via a temp file in the
// same directory, so a partial download is never visible as a cache hit.
func (h *Handler) fill(ctx context.Context, name, local string) error {
	u := *h.upstream
	u.Path = path.Join(u.Path, name)

	resp, err := h.get(ctx, u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		status := resp.StatusCode
		if status != http.StatusNotFound {
			status = http.StatusBadGateway
		}

		return &upstreamStatusError{url: u.String(), status: status}
	}

	// go.dev/dl serves a 200 HTML "redirecting to documentation" page for
	// names it does not know instead of a 404. Never cache that as an archive.
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		return &upstreamStatusError{url: u.String(), status: http.StatusNotFound}
	}

	tmp, err := os.CreateTemp(h.cacheDir, "."+name+".*.part")
	if err != nil {
		return fmt.Errorf("godl: creating temp file: %w", err)
	}

	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	written, err := io.Copy(tmp, resp.Body)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}

	if err != nil {
		return fmt.Errorf("godl: downloading %s: %w", u.String(), err)
	}

	if resp.ContentLength >= 0 && written != resp.ContentLength {
		return fmt.Errorf("godl: downloading %s: got %d bytes, want %d", u.String(), written, resp.ContentLength)
	}

	if err := os.Rename(tmpName, local); err != nil {
		return fmt.Errorf("godl: committing %s to cache: %w", name, err)
	}

	return nil
}

func (h *Handler) get(ctx context.Context, rawURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("godl: building request for %s: %w", rawURL, err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("godl: fetching %s: %w", rawURL, err)
	}

	return resp, nil
}
