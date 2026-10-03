// Package godl serves the Go release download protocol using Athens storage.
package godl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
)

var archiveRE = regexp.MustCompile(`^go[0-9][A-Za-z0-9.\-]*\.(tar\.gz|zip|msi|pkg)$`)

// Handler adapts the Go release HTTP protocol to the release service.
type Handler struct{ service *Service }

func New(opts Options) (*Handler, error) {
	service, err := NewService(opts)
	if err != nil {
		return nil, err
	}
	return &Handler{service: service}, nil
}

func Disabled() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Go toolchain downloads are disabled: set GoDownloadEnabled or GoDownloadURL (ATHENS_GO_DOWNLOAD_URL)", http.StatusUnprocessableEntity)
	})
}

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
	case archiveRE.MatchString(name):
		h.serveArchive(w, r, name)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) serveListing(w http.ResponseWriter, r *http.Request) {
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "invalid listing query", http.StatusBadRequest)
		return
	}
	for key, values := range q {
		if len(values) != 1 || (key != "mode" && key != "include") || (key == "include" && values[0] != "all" && values[0] != "") {
			http.Error(w, "unsupported listing query", http.StatusBadRequest)
			return
		}
	}
	releases, modified, err := h.service.Listing(r.Context(), q.Get("include") == "all")
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	if releases == nil {
		releases = []storage.Release{}
	}
	body, err := json.Marshal(releases)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprint(len(body)))
	// Discovery can change after archive publication/deletion. Do not let a
	// downstream cache advertise artifacts no longer available in this mirror.
	w.Header().Set("Cache-Control", "no-cache")
	if !modified.IsZero() {
		w.Header().Set("Last-Modified", modified.UTC().Format(http.TimeFormat))
	}
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(body)
	}
}

func (h *Handler) serveArchive(w http.ResponseWriter, r *http.Request, name string) {
	info, body, err := h.service.Archive(r.Context(), name)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	reader := &archiveSeeker{ctx: r.Context(), store: h.service.opts.Storage, source: h.service.opts.Source, name: name, info: info, body: body}
	defer reader.Close()
	w.Header().Set("ETag", `"sha256-`+strings.ToLower(info.SHA256)+`"`)
	w.Header().Set("Cache-Control", h.service.opts.CacheControl)
	http.ServeContent(w, r, name, info.SavedAt, reader)
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, errors.KindRedirect) && h.service.opts.RedirectURL != nil {
		u := *h.service.opts.RedirectURL
		u.Path = strings.TrimRight(u.Path, "/") + r.URL.Path
		u.RawQuery = r.URL.RawQuery
		http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
		return
	}
	status := errors.Kind(err)
	if status == 0 {
		status = http.StatusInternalServerError
	}
	http.Error(w, err.Error(), status)
}

// archiveSeeker provides ServeContent semantics over streaming backends. Seeks
// reopen the stored archive and discard only the requested prefix; they never
// refetch upstream or load an entire archive into memory.
type archiveSeeker struct {
	//nolint:containedctx // ServeContent requires io.ReadSeeker; reopening uses the HTTP request context.
	ctx          context.Context
	store        storage.ArchiveReader
	source, name string
	info         storage.ArchiveInfo
	body         storage.SizeReadCloser
	position     int64
	needsOpen    bool
}

func (s *archiveSeeker) Read(p []byte) (int, error) {
	if s.position >= s.info.Size {
		return 0, io.EOF
	}
	if s.needsOpen {
		info, body, err := s.store.Archive(s.ctx, s.source, s.name)
		if err != nil {
			return 0, err
		}
		if info.SHA256 != s.info.SHA256 || info.Size != s.info.Size {
			_ = body.Close()
			return 0, fmt.Errorf("archive changed during response")
		}
		s.body = body
		s.needsOpen = false
		if _, err := io.CopyN(io.Discard, s.body, s.position); err != nil {
			return 0, err
		}
	}
	n, err := s.body.Read(p)
	s.position += int64(n)
	return n, err
}

func (s *archiveSeeker) Seek(offset int64, whence int) (int64, error) {
	position := offset
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		position += s.position
	case io.SeekEnd:
		position += s.info.Size
	default:
		return 0, fmt.Errorf("invalid seek origin")
	}
	if position < 0 || position > s.info.Size {
		return 0, fmt.Errorf("invalid archive offset")
	}
	if position != s.position {
		if s.body != nil {
			_ = s.body.Close()
			s.body = nil
		}
		s.position = position
		s.needsOpen = true
	}
	return position, nil
}

func (s *archiveSeeker) Close() error {
	if s.body != nil {
		return s.body.Close()
	}
	return nil
}
