package events

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gomods/athens/pkg/requestid"
)

// NewServer returns an http.Handler that parses
func NewServer(h Hook) http.Handler {
	return &server{h}
}

type server struct {
	h Hook
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	ctx := r.Context()
	ctx = requestid.SetInContext(ctx, r.Header.Get(requestid.HeaderKey))
	var err error
	switch event := r.Header.Get(HeaderKey); event {
	case Ping.String():
		err = s.h.Ping(ctx)
	case Stashed.String():
		var body StashedEvent
		err = json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			break
		}
		err = s.h.Stashed(ctx, body.Module, body.Version)
	default:
		err = fmt.Errorf("unknown event: %q", event)
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
