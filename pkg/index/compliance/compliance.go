package compliance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/index"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/technosophos/moniker"
)

func RunTests(t *testing.T, indexer index.Indexer, clearIndex func() error) {
	if err := clearIndex(); err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		name    string
		desc    string
		limit   int
		preTest func(t *testing.T) ([]*index.Line, time.Time)
	}{
		{
			name:    "empty",
			desc:    "an empty index should return an empty slice",
			preTest: func(t *testing.T) ([]*index.Line, time.Time) { return []*index.Line{}, time.Time{} },
			limit:   2000,
		},
		{
			name: "happy path",
			desc: "given 10 modules, return all of them in correct order",
			preTest: func(t *testing.T) ([]*index.Line, time.Time) {
				return seed(t, indexer, 10), time.Time{}
			},
			limit: 2000,
		},
		{
			name: "respect the limit",
			desc: "givn 10 modules and a 'limit' of 5, only return the first five lines",
			preTest: func(t *testing.T) ([]*index.Line, time.Time) {
				lines := seed(t, indexer, 10)
				return lines[0:5], time.Time{}
			},
			limit: 5,
		},
		{
			name: "respect the time",
			desc: "given 10 modules, 'since' should filter out the ones that came before it",
			preTest: func(t *testing.T) ([]*index.Line, time.Time) {
				err := indexer.Index(context.Background(), "tobeignored", "v1.2.3")
				if err != nil {
					t.Fatal(err)
				}
				time.Sleep(50 * time.Millisecond)
				now := time.Now()
				lines := seed(t, indexer, 5)
				return lines, now
			},
			limit: 2000,
		},
		{
			name: "ignore the past",
			desc: "no line should be returned if 'since' is after all of the indexed modules",
			preTest: func(t *testing.T) ([]*index.Line, time.Time) {
				seed(t, indexer, 5)
				time.Sleep(50 * time.Millisecond)
				return []*index.Line{}, time.Now()
			},
			limit: 2000,
		},
		{
			name: "no limit no line",
			desc: "if limit is set to zero, then nothing should be returned",
			preTest: func(t *testing.T) ([]*index.Line, time.Time) {
				seed(t, indexer, 5)
				return []*index.Line{}, time.Time{}
			},
			limit: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.desc)
			t.Cleanup(func() {
				if err := clearIndex(); err != nil {
					t.Fatal(err)
				}
			})
			expected, since := tc.preTest(t)
			given, err := indexer.Lines(context.Background(), since, tc.limit)
			if err != nil {
				t.Fatal(err)
			}
			opts := cmpopts.IgnoreFields(index.Line{}, "Timestamp")
			if !cmp.Equal(given, expected, opts) {
				t.Fatal(cmp.Diff(expected, given, opts))
			}
		})
	}
}

func seed(t *testing.T, indexer index.Indexer, num int) []*index.Line {
	lines := []*index.Line{}
	t.Helper()
	for i := 0; i < num; i++ {
		mod := moniker.New().NameSep("_")
		ver := fmt.Sprintf("%d.0.0", i)
		err := indexer.Index(context.Background(), mod, ver)
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, &index.Line{Module: mod, Version: ver})
	}
	return lines
}
