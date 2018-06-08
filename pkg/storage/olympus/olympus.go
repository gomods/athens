package olympus

import (
	"net/http"
	"time"

	"github.com/gomods/athens/pkg/eventlog"
)

// ModuleStore represents a mongo backed storage backend.
type ModuleStore struct {
	url    string
	client http.Client
}

// NewStorage returns a remote Olympus store
func NewStorage(url string) *ModuleStore {
	client := http.Client{
		Timeout: 180 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return &eventlog.ErrUseNewOlympus{Endpoint: req.URL.String()}
		},
	}
	return &ModuleStore{url: url, client: client}
}
