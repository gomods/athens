package actions

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/index"
)

var indexHandlerTests = []struct {
	name  string
	desc  string
	lines []*index.Line
	err   error
	limit string
	since string
	code  int
}{
	{
		name: "happy path",
		desc: "given no params and 1 line, the handler should return 200 along with the index line",
		lines: []*index.Line{
			{
				Path:      "github.com/pkg/errors",
				Version:   "v0.9.1",
				Timestamp: time.Now(),
			},
		},
		code: 200,
	},
	{
		name: "valid limit",
		desc: "given a valid limit number, the handler should return 200",
		lines: []*index.Line{
			{
				Path:      "github.com/pkg/errors",
				Version:   "v0.9.1",
				Timestamp: time.Now(),
			},
		},
		limit: "1",
		code:  200,
	},
	{
		name: "valid since",
		desc: "given a valid since string, the handler should return 200",
		lines: []*index.Line{
			{
				Path:      "github.com/pkg/errors",
				Version:   "v0.9.1",
				Timestamp: time.Now(),
			},
		},
		limit: "1",
		since: time.Now().Add(-time.Hour).Format(time.RFC3339),
		code:  200,
	},
	{
		name:  "invalid limit",
		desc:  "a limit query param must be a valid integer",
		limit: "im not an integer",
		code:  400,
	},
	{
		name:  "limit too low",
		desc:  "a limit query cannot be a negative number",
		limit: "-1",
		code:  400,
	},
	{
		name:  "invalid zero limit",
		desc:  "a limit cannot be 0",
		limit: "0",
		code:  400,
	},
	{
		name:  "invalid since",
		desc:  "since must be a valid RFC3339 format",
		since: time.Now().Format(time.RFC822),
		code:  400,
	},
	{
		name: "index error",
		desc: "given an underlying index error, the handler must return 500",
		err:  fmt.Errorf("internal error"),
		code: 500,
	},
}

func TestIndexHandler(t *testing.T) {
	for _, tc := range indexHandlerTests {
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.desc)
			req := httptest.NewRequest("GET", "/index", nil)
			q := url.Values{}
			q.Set("limit", tc.limit)
			q.Set("since", tc.since)
			req.URL.RawQuery = q.Encode()
			w := httptest.NewRecorder()
			mi := &mockIndexer{lines: tc.lines, err: tc.err}
			handler := indexHandler(mi)
			handler(w, req)
			if w.Code != tc.code {
				t.Fatalf("expected response code to be %d but got %d", tc.code, w.Code)
			}
		})
	}
}

type mockIndexer struct {
	index.Indexer

	lines []*index.Line
	err   error
}

func (mi *mockIndexer) Lines(ctx context.Context, since time.Time, limit int) ([]*index.Line, error) {
	return mi.lines, mi.err
}
