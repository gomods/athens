package external

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gorilla/mux"
)

const (
	releasePrefix     = "/toolchains/v1"
	archiveInfoHeader = "Athens-Archive-Info"
)

// CheckToolchainStorage checks remote capability before Athens enables /dl/.
func (s *service) CheckToolchainStorage(ctx context.Context) error {
	resp, err := s.releaseRequest(ctx, http.MethodGet, "", "", nil, nil)
	if err != nil {
		return fmt.Errorf("external storage must support the Athens toolchains/v1 API: %w", err)
	}
	defer resp.Body.Close()
	var version struct{ Version int }
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return err
	}
	if version.Version != 1 {
		return fmt.Errorf("unsupported external toolchain storage version %d", version.Version)
	}
	return nil
}

func (s *service) releaseRequest(ctx context.Context, method, endpoint, source string, info *storage.ArchiveInfo, body io.Reader) (*http.Response, error) {
	u := s.url + releasePrefix + endpoint
	if source != "" {
		u += "?" + url.Values{"source": {source}}.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}
	if info != nil {
		metadata, err := json.Marshal(info)
		if err != nil {
			return nil, err
		}
		req.Header.Set(archiveInfoHeader, base64.RawURLEncoding.EncodeToString(metadata))
	}
	resp, err := s.c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		_ = resp.Body.Close()
		return nil, errors.E("external.Toolchain", fmt.Sprintf("external storage returned %d: %s", resp.StatusCode, message), resp.StatusCode)
	}
	return resp, nil
}

func (s *service) Archive(ctx context.Context, source, filename string) (storage.ArchiveInfo, storage.SizeReadCloser, error) {
	var info storage.ArchiveInfo
	resp, err := s.releaseRequest(ctx, http.MethodGet, "/archive/"+url.PathEscape(filename), source, nil, nil)
	if err != nil {
		return info, nil, err
	}
	if err := decodeArchiveInfo(resp.Header.Get(archiveInfoHeader), &info); err != nil {
		_ = resp.Body.Close()
		return info, nil, err
	}
	if resp.ContentLength != info.Size {
		_ = resp.Body.Close()
		return info, nil, fmt.Errorf("external archive length does not match metadata")
	}
	return info, storage.NewSizer(resp.Body, info.Size), nil
}

func (s *service) SaveArchive(ctx context.Context, source string, info storage.ArchiveInfo, body io.Reader) error {
	resp, err := s.releaseRequest(ctx, http.MethodPost, "/archive/"+url.PathEscape(info.Filename), source, &info, body)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (s *service) ListArchives(ctx context.Context, source string) ([]storage.ArchiveInfo, error) {
	resp, err := s.releaseRequest(ctx, http.MethodGet, "/archives", source, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var infos []storage.ArchiveInfo
	err = json.NewDecoder(resp.Body).Decode(&infos)
	return infos, err
}

func (s *service) DeleteArchive(ctx context.Context, source, filename string) error {
	resp, err := s.releaseRequest(ctx, http.MethodDelete, "/archive/"+url.PathEscape(filename), source, nil, nil)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (s *service) Releases(ctx context.Context, source string) (*storage.ReleaseList, error) {
	resp, err := s.releaseRequest(ctx, http.MethodGet, "/releases", source, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var list storage.ReleaseList
	err = json.NewDecoder(io.LimitReader(resp.Body, 16<<20)).Decode(&list)
	return &list, err
}

func (s *service) SaveReleases(ctx context.Context, source string, list *storage.ReleaseList) error {
	body, err := json.Marshal(list)
	if err != nil {
		return err
	}
	resp, err := s.releaseRequest(ctx, http.MethodPost, "/releases", source, nil, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func decodeArchiveInfo(encoded string, info *storage.ArchiveInfo) error {
	body, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, info)
}

func registerToolchainStorage(r *mux.Router, backend storage.Backend) {
	store, ok := backend.(storage.ToolchainStorage)
	if !ok {
		return
	}
	r.HandleFunc(releasePrefix, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"Version":1}`)
	}).Methods(http.MethodGet)
	r.HandleFunc(releasePrefix+"/archives", func(w http.ResponseWriter, r *http.Request) {
		infos, err := store.ListArchives(r.Context(), r.URL.Query().Get("source"))
		if err != nil {
			http.Error(w, err.Error(), errors.Kind(err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(infos)
	}).Methods(http.MethodGet)
	r.HandleFunc(releasePrefix+"/archive/{filename}", func(w http.ResponseWriter, r *http.Request) {
		source, name := r.URL.Query().Get("source"), mux.Vars(r)["filename"]
		var err error
		switch r.Method {
		case http.MethodGet:
			var info storage.ArchiveInfo
			var body storage.SizeReadCloser
			info, body, err = store.Archive(r.Context(), source, name)
			if err == nil {
				defer body.Close()
				metadata, e := json.Marshal(info)
				if e != nil {
					err = e
					break
				}
				w.Header().Set(archiveInfoHeader, base64.RawURLEncoding.EncodeToString(metadata))
				w.Header().Set("Content-Length", fmt.Sprint(body.Size()))
				_, _ = io.Copy(w, body)
				return
			}
		case http.MethodPost:
			var info storage.ArchiveInfo
			err = decodeArchiveInfo(r.Header.Get(archiveInfoHeader), &info)
			if err != nil {
				err = errors.E("external.SaveArchive", err, errors.KindBadRequest)
			}
			if err == nil && info.Filename != name {
				err = errors.E("external.SaveArchive", "filename mismatch", errors.KindBadRequest)
			}
			if err == nil {
				r.Body = http.MaxBytesReader(w, r.Body, (1<<30)+1)
				err = store.SaveArchive(r.Context(), source, info, r.Body)
			}
		case http.MethodDelete:
			err = store.DeleteArchive(r.Context(), source, name)
		}
		if err != nil {
			http.Error(w, err.Error(), errors.Kind(err))
		}
	}).Methods(http.MethodGet, http.MethodPost, http.MethodDelete)
	r.HandleFunc(releasePrefix+"/releases", func(w http.ResponseWriter, r *http.Request) {
		source := r.URL.Query().Get("source")
		var err error
		if r.Method == http.MethodGet {
			var list *storage.ReleaseList
			list, err = store.Releases(r.Context(), source)
			if err == nil {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(list)
				return
			}
		} else {
			var list storage.ReleaseList
			r.Body = http.MaxBytesReader(w, r.Body, 16<<20)
			err = json.NewDecoder(r.Body).Decode(&list)
			if err != nil {
				err = errors.E("external.SaveReleases", err, errors.KindBadRequest)
			}
			if err == nil {
				err = store.SaveReleases(r.Context(), source, &list)
			}
		}
		if err != nil {
			http.Error(w, err.Error(), errors.Kind(err))
		}
	}).Methods(http.MethodGet, http.MethodPost)
}
