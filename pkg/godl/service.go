package godl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"golang.org/x/mod/semver"
	"golang.org/x/sync/singleflight"
)

const DefaultListingTTL = 2 * time.Hour

// Options configures release storage, upstream access and miss policy.
// Module-path download overrides do not apply to release archives.
type Options struct {
	Storage storage.ToolchainStorage
	//nolint:containedctx // Constructor option supplies the service lifetime, not request state.
	Context      context.Context
	CacheControl string
	Upstream     *url.URL
	Source       string
	Client       *http.Client
	ListingTTL   time.Duration
	NetworkMode  string
	DownloadMode mode.Mode
	RedirectURL  *url.URL
	Workers      int
	Timeout      time.Duration
}

// Service downloads verified releases into the configured Athens storage.
// It keeps upstream discovery separate from the locally available inventory.
type Service struct {
	opts Options
	//nolint:containedctx // Service-owned lifetime cancels detached background work on shutdown.
	ctx        context.Context
	cancel     context.CancelFunc
	group      singleflight.Group
	workers    chan struct{}
	background chan struct{}
	active     sync.Map
	refreshMu  sync.Mutex
	retryAt    time.Time
	refreshErr error
}

func NewService(opts Options) (*Service, error) {
	if opts.Storage == nil || opts.Source == "" {
		return nil, fmt.Errorf("godl: storage and distribution identity are required")
	}
	if opts.NetworkMode == "" {
		opts.NetworkMode = download.Strict
	}
	switch opts.NetworkMode {
	case download.Strict, download.Fallback, download.Offline:
	default:
		return nil, fmt.Errorf("godl: unknown network mode %q", opts.NetworkMode)
	}
	if opts.DownloadMode == "" {
		opts.DownloadMode = mode.Sync
	}
	switch opts.DownloadMode {
	case mode.Sync, mode.None, mode.Async, mode.Redirect, mode.AsyncRedirect:
	default:
		return nil, fmt.Errorf("godl: unknown download mode %q", opts.DownloadMode)
	}
	var err error
	if opts.Upstream, err = releaseURL(opts.Upstream); err != nil {
		return nil, err
	}
	if opts.Upstream == nil && opts.NetworkMode != download.Offline && opts.DownloadMode != mode.None {
		return nil, fmt.Errorf("godl: upstream URL is required for online downloads")
	}
	if opts.RedirectURL == nil {
		opts.RedirectURL = opts.Upstream
	}
	if opts.RedirectURL, err = releaseURL(opts.RedirectURL); err != nil {
		return nil, err
	}
	if opts.Client == nil {
		opts.Client = http.DefaultClient
	}
	if opts.ListingTTL <= 0 {
		opts.ListingTTL = DefaultListingTTL
	}
	if opts.Workers <= 0 {
		opts.Workers = 4
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 5 * time.Minute
	}
	if opts.Context == nil {
		opts.Context = context.Background()
	}
	if opts.CacheControl == "" {
		opts.CacheControl = "private, no-cache"
	}
	lifetime, cancel := context.WithCancel(opts.Context)
	return &Service{opts: opts, ctx: lifetime, cancel: cancel, workers: make(chan struct{}, opts.Workers), background: make(chan struct{}, opts.Workers)}, nil
}

func (s *Service) storageOnly() bool {
	return s.opts.NetworkMode == download.Offline || s.opts.DownloadMode == mode.None
}

// Listing returns discovery metadata, or only committed local files when
// upstream access is disabled or fallback discovery is necessary.
func (s *Service) Listing(ctx context.Context, all bool) ([]storage.Release, time.Time, error) {
	if s.storageOnly() {
		return s.inventory(ctx, all)
	}
	list, err := s.opts.Storage.Releases(ctx, s.opts.Source)
	if err != nil && !errors.Is(err, errors.KindNotFound) {
		return nil, time.Time{}, err
	}
	if list != nil && time.Since(list.FetchedAt) < s.opts.ListingTTL {
		return selectReleases(list.Releases, all), list.FetchedAt, nil
	}
	result := s.group.DoChan("listing", func() (any, error) {
		s.refreshMu.Lock()
		defer s.refreshMu.Unlock()
		if time.Now().Before(s.retryAt) {
			return nil, s.refreshErr
		}
		fresh, err := s.refresh(ctx)
		if err != nil {
			s.retryAt = time.Now().Add(10 * time.Second)
			s.refreshErr = err
		}
		return fresh, err
	})
	select {
	case <-ctx.Done():
		return nil, time.Time{}, ctx.Err()
	case r := <-result:
		if r.Err != nil {
			if s.opts.NetworkMode == download.Fallback {
				return s.inventory(ctx, all)
			}
			return nil, time.Time{}, r.Err
		}
		fresh := r.Val.(*storage.ReleaseList)
		return selectReleases(fresh.Releases, all), fresh.FetchedAt, nil
	}
}

func (s *Service) refresh(caller context.Context) (*storage.ReleaseList, error) {
	ctx, cancel := s.workContext(caller)
	defer cancel()
	fetchedAt := time.Now().UTC()
	u := *s.opts.Upstream
	u.Path += "/"
	u.RawQuery = "mode=json&include=all"
	resp, err := s.get(ctx, u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, upstreamError(resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, (16<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(body) > 16<<20 {
		return nil, errors.E("godl.Listing", "listing exceeds size limit", http.StatusBadGateway)
	}
	var releases []storage.Release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, errors.E("godl.Listing", err, http.StatusBadGateway)
	}
	releases, err = verifiedReleases(releases)
	if err != nil {
		return nil, err
	}
	list := &storage.ReleaseList{Releases: releases, FetchedAt: fetchedAt}
	if err := s.opts.Storage.SaveReleases(ctx, s.opts.Source, list); err != nil {
		return nil, err
	}
	return list, nil
}

func (s *Service) inventory(ctx context.Context, all bool) ([]storage.Release, time.Time, error) {
	files, err := s.opts.Storage.ListArchives(ctx, s.opts.Source)
	if err != nil {
		return nil, time.Time{}, err
	}
	releases := map[string]*storage.Release{}
	var modified time.Time
	for _, file := range files {
		if file.Kind != "archive" && file.Kind != "source" {
			continue
		}
		r := releases[file.Version]
		if r == nil {
			r = &storage.Release{Version: file.Version, Stable: file.Stable, Files: []storage.ReleaseFile{}}
			releases[file.Version] = r
		}
		r.Files = append(r.Files, file.ReleaseFile)
		if file.SavedAt.After(modified) {
			modified = file.SavedAt
		}
	}
	list := make([]storage.Release, 0, len(releases))
	for _, release := range releases {
		list = append(list, *release)
	}
	return selectReleases(list, all), modified, nil
}

func goSemver(version string) string {
	v := strings.TrimPrefix(version, "go")
	v = strings.Replace(v, "beta", "-beta.", 1)
	v = strings.Replace(v, "rc", "-rc.", 1)
	p := strings.SplitN(v, "-", 2)
	for strings.Count(p[0], ".") < 2 {
		p[0] += ".0"
	}
	return "v" + strings.Join(p, "-")
}

func selectReleases(releases []storage.Release, all bool) []storage.Release {
	result := slices.Clone(releases)
	slices.SortFunc(result, func(a, b storage.Release) int { return semver.Compare(goSemver(b.Version), goSemver(a.Version)) })
	if all {
		return result
	}
	// Match go.dev's default listing: newest patch for the two supported stable
	// release series. Prereleases are available only with include=all.
	selected := make([]storage.Release, 0)
	series := map[string]bool{}
	for _, r := range result {
		if !r.Stable {
			continue
		}
		minor := semver.MajorMinor(goSemver(r.Version))
		if series[minor] || len(series) >= 2 {
			continue
		}
		series[minor] = true
		selected = append(selected, r)
	}
	return selected
}

// Archive serves committed storage hits before applying download miss policy.
func (s *Service) Archive(ctx context.Context, name string) (storage.ArchiveInfo, storage.SizeReadCloser, error) {
	info, body, err := s.opts.Storage.Archive(ctx, s.opts.Source, name)
	if err == nil {
		observ.RecordCacheLookup(ctx, "hit", "toolchain_archive")
		return info, body, nil
	}
	if !errors.Is(err, errors.KindNotFound) {
		return info, nil, err
	}
	observ.RecordCacheLookup(ctx, "miss", "toolchain_archive")
	if s.storageOnly() {
		return info, nil, errors.E("godl.Archive", "archive is not stored", errors.KindNotFound)
	}
	switch s.opts.DownloadMode {
	case mode.Redirect:
		return info, nil, errors.E("godl.Archive", "release redirect", errors.KindRedirect)
	case mode.Async, mode.AsyncRedirect:
		if err := s.schedule(ctx, name); err != nil {
			return info, nil, err
		}
		kind := errors.KindNotFound
		if s.opts.DownloadMode == mode.AsyncRedirect {
			kind = errors.KindRedirect
		}
		return info, nil, errors.E("godl.Archive", "archive fetch scheduled", kind)
	}
	if err := s.download(ctx, name); err != nil {
		return info, nil, err
	}
	return s.opts.Storage.Archive(ctx, s.opts.Source, name)
}

func (s *Service) download(ctx context.Context, name string) error {
	result := s.group.DoChan("archive:"+name, func() (any, error) {
		work, cancel := s.workContext(ctx)
		defer cancel()
		select {
		case s.workers <- struct{}{}:
			defer func() { <-s.workers }()
		case <-work.Done():
			return nil, work.Err()
		}
		return nil, s.fill(work, name)
	})
	select {
	case <-ctx.Done():
		return ctx.Err()
	case r := <-result:
		return r.Err
	}
}

func (s *Service) schedule(ctx context.Context, name string) error {
	if _, loaded := s.active.LoadOrStore(name, true); loaded {
		return nil
	}
	select {
	case s.background <- struct{}{}:
	default:
		s.active.Delete(name)
		return errors.E("godl.Archive", "background download capacity reached", errors.KindRateLimit)
	}
	go func() {
		defer s.active.Delete(name)
		defer func() { <-s.background }()
		work, cancel := s.workContext(ctx)
		defer cancel()
		if err := s.download(work, name); err != nil && work.Err() == nil {
			log.EntryFromContext(work).SystemErr(errors.E("godl.Background", err))
		}
	}()
	return nil
}

func (s *Service) fill(ctx context.Context, name string) error {
	_, existing, err := s.opts.Storage.Archive(ctx, s.opts.Source, name)
	if err == nil {
		return existing.Close()
	}
	if !errors.Is(err, errors.KindNotFound) {
		return err
	}
	listing, _, err := s.Listing(ctx, true)
	if err != nil {
		return err
	}
	var info storage.ArchiveInfo
	found := false
	for _, release := range listing {
		for _, file := range release.Files {
			if file.Filename == name {
				info = storage.ArchiveInfo{ReleaseFile: file, Stable: release.Stable}
				found = true
			}
		}
	}
	if !found {
		return errors.E("godl.Archive", "archive absent from verified release listing", errors.KindNotFound)
	}
	u := *s.opts.Upstream
	u.Path = path.Join(u.Path, name)
	resp, err := s.get(ctx, u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return upstreamError(resp.StatusCode)
	}
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		return errors.E("godl.Archive", "upstream returned HTML", errors.KindNotFound)
	}
	if err := s.opts.Storage.SaveArchive(ctx, s.opts.Source, info, upstreamReader{resp.Body}); err != nil {
		if errors.Is(err, errors.KindBadRequest) {
			return errors.E("godl.Archive", err, http.StatusBadGateway)
		}
		return err
	}
	return nil
}

func (s *Service) get(ctx context.Context, rawURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.opts.Client.Do(req)
	if err != nil {
		return nil, errors.E("godl.Upstream", err, http.StatusBadGateway)
	}
	return resp, nil
}

func upstreamError(status int) error {
	kind := http.StatusBadGateway
	if status == http.StatusNotFound {
		kind = errors.KindNotFound
	}
	return errors.E("godl.Upstream", fmt.Sprintf("release upstream returned %d", status), kind)
}

// upstreamReader distinguishes transport failures from backend write failures.
type upstreamReader struct{ io.Reader }

func (r upstreamReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err != nil && err != io.EOF {
		return n, errors.E("godl.Upstream", err, http.StatusBadGateway)
	}
	return n, err
}

// Close cancels detached refreshes and background downloads.
func (s *Service) Close() { s.cancel() }

func (s *Service) workContext(caller context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(caller), s.opts.Timeout)
	stop := context.AfterFunc(s.ctx, cancel)
	if s.ctx.Err() != nil {
		cancel()
	}
	return ctx, func() { stop(); cancel() }
}

func releaseURL(base *url.URL) (*url.URL, error) {
	if base == nil {
		return nil, nil
	}
	u := *base
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return nil, fmt.Errorf("godl: release base must be an HTTP(S) URL without credentials, query or fragment")
	}
	u.Path = strings.TrimRight(u.Path, "/")
	return &u, nil
}

func verifiedReleases(releases []storage.Release) ([]storage.Release, error) {
	seen := map[string]storage.ArchiveInfo{}
	result := make([]storage.Release, 0, len(releases))
	for _, release := range releases {
		if !strings.HasPrefix(release.Version, "go") || !semver.IsValid(goSemver(release.Version)) {
			return nil, errors.E("godl.Listing", "invalid Go release version", http.StatusBadGateway)
		}
		files := make([]storage.ReleaseFile, 0, len(release.Files))
		for _, file := range release.Files {
			// Historical go.dev entries without a checksum or size cannot be
			// verified. Omit them rather than advertise a download we cannot serve.
			if file.SHA256 == "" || file.Size == 0 {
				continue
			}
			if err := file.Validate(); err != nil || file.Version != release.Version {
				return nil, errors.E("godl.Listing", "invalid archive metadata", http.StatusBadGateway)
			}
			file.SHA256 = strings.ToLower(file.SHA256)
			metadata := storage.ArchiveInfo{ReleaseFile: file, Stable: release.Stable}
			if previous, ok := seen[file.Filename]; ok {
				if previous != metadata {
					return nil, errors.E("godl.Listing", "conflicting archive metadata", http.StatusBadGateway)
				}
				continue
			}
			seen[file.Filename] = metadata
			files = append(files, file)
		}
		if len(files) > 0 {
			// setup-go selects by platform without checking kind. Keep usable
			// archive files ahead of installers even if a mirror orders them differently.
			slices.SortStableFunc(files, func(a, b storage.ReleaseFile) int {
				if (a.Kind == "installer") == (b.Kind == "installer") {
					return 0
				}
				if a.Kind == "installer" {
					return 1
				}
				return -1
			})
			release.Files = files
			result = append(result, release)
		}
	}
	return result, nil
}
