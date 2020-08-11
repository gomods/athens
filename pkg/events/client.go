package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gomods/athens/pkg/build"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/requestid"
)

// NewClient returns a new http service
func NewClient(url string, c *http.Client) Hook {
	if c == nil {
		c = http.DefaultClient
	}
	return &service{url, c}
}

type service struct {
	url string
	c   *http.Client
}

func (s *service) Ping(ctx context.Context) error {
	const op errors.Op = "events.Ping"
	return s.sendEvent(ctx, op, Ping, PingEvent{BaseEvent: BaseEvent{
		Event:   Ping.String(),
		Version: build.Data().Version,
	}})
}

func (s *service) Stashed(ctx context.Context, mod, ver string) error {
	const op errors.Op = "events.Stashed"
	return s.sendEvent(ctx, op, Stashed, StashedEvent{
		BaseEvent: BaseEvent{
			Event:   Stashed.String(),
			Version: build.Data().Version,
		},
		Module:  mod,
		Version: ver,
	})
}

func (s *service) sendEvent(ctx context.Context, op errors.Op, event Type, payload interface{}) error {
	req, err := s.getRequest(ctx, event, payload)
	if err != nil {
		return errors.E(op, err)
	}
	resp, err := s.c.Do(req)
	if err != nil {
		return errors.E(op, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.E(op, fmt.Errorf("event backend returned non-200 code: %d - body: %s", resp.StatusCode, body))
	}
	return nil
}

func (s *service) getRequest(ctx context.Context, event Type, payload interface{}) (*http.Request, error) {
	const op errors.Op = "events.getRequest"
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(payload)
	if err != nil {
		return nil, errors.E(op, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, &buf)
	if err != nil {
		return nil, errors.E(op, err)
	}
	req.Header.Set(HeaderKey, event.String())
	req.Header.Set(requestid.HeaderKey, requestid.FromContext(ctx))
	return req, nil
}
