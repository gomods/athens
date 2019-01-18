package http

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// ModuleStore represents an HTTP-backed storage backend.
type ModuleStore struct {
	url      string // base url
	username string // http basic auth username
	password string // http basic auth password
	timeout  time.Duration
	client   *http.Client
}

// New returns a new HTTP backed storage
// that satisfies the Backend interface.
func New(conf *config.HTTPConfig, timeout time.Duration) (*ModuleStore, error) {
	const op errors.Op = "http.NewStorage"
	if conf == nil {
		return nil, errors.E(op, "No HTTP Configuration provided")
	}
	ms := &ModuleStore{url: conf.BaseURL, username: conf.Username, password: conf.Password, timeout: timeout, client: http.DefaultClient}

	err := ms.connect()
	if err != nil {
		return nil, errors.E(op, err)
	}

	return ms, nil
}

func (m *ModuleStore) connect() error {
	const op errors.Op = "http.connect"

	// I guess just GET the base URL and see if it 401's?
	req, _ := http.NewRequest(http.MethodGet, m.url, nil)
	req.SetBasicAuth(m.username, m.password)
	resp, err := m.client.Do(req)
	if err != nil {
		return errors.E(op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.E(op, resp.Status)
	}

	return nil

}

// moduleRoot determines the root URL for a module.
func (m *ModuleStore) moduleRoot(module string) string {
	return m.url + "/" + module + "/@v/"
}

func (s *ModuleStore) doRequest(ctx context.Context, req *http.Request, expectedStatus int) (io.ReadCloser, error) {
	const op errors.Op = "http.doRequest"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	// set credentials if configured with any
	if len(s.username) > 0 || len(s.password) > 0 {
		req.SetBasicAuth(s.username, s.password)
	}

	resp, err := s.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.E(op, err)
	}

	if resp.StatusCode != expectedStatus {
		kind := errors.KindUnexpected
		if resp.StatusCode == http.StatusNotFound {
			kind = errors.KindNotFound
		}
		io.Copy(ioutil.Discard, resp.Body)
		return nil, errors.E(op, kind)
	}

	return resp.Body, nil

}

// fetchFile downloads a file from a remote URL.
func (s *ModuleStore) fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	const op errors.Op = "http.fetchFile"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return s.doRequest(ctx, req, http.StatusOK)

}

func (s *ModuleStore) upload(ctx context.Context, path, contentType string, stream io.Reader) error {
	const op errors.Op = "http.upload"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	req, _ := http.NewRequest(http.MethodPut, s.url+"/"+path, stream)
	req.Header.Set("Content-Type", contentType)

	body, err := s.doRequest(ctx, req, http.StatusCreated)
	if err != nil {
		return errors.E(op, err)
	}

	// we don't actually care what the body is but we'll throw it away
	// so we can reuse the underlying connection
	io.Copy(ioutil.Discard, body)

	return nil
}

func (s *ModuleStore) remove(ctx context.Context, path string) error {
	const op errors.Op = "http.deleteFile"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	req, _ := http.NewRequest(http.MethodDelete, s.url+"/"+path, nil)
	body, err := s.doRequest(ctx, req, http.StatusOK)
	if err != nil {
		return errors.E(op, err)
	}

	// we don't actually care what the body is but we'll throw it away
	// so we can reuse the underlying connection
	io.Copy(ioutil.Discard, body)

	return nil
}
