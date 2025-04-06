package log

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type input struct {
	name          string
	cloudProvider string
	format        string
	level         slog.Level
	fields        map[string]any
	logFunc       func(e Entry) time.Time
	output        string
}

var testCases = []input{
	{
		name:          "gcp_debug",
		cloudProvider: "GCP",
		level:         slog.LevelDebug,
		fields:        map[string]any{},
		logFunc: func(e Entry) time.Time {
			t := time.Now()
			e.Infof("info message")
			return t
		},
		output: `{"message":"info message","severity":"info","timestamp":"%v"}` + "\n",
	},
	{
		name:          "gcp_error",
		cloudProvider: "GCP",
		level:         slog.LevelDebug,
		fields:        map[string]any{},
		logFunc: func(e Entry) time.Time {
			t := time.Now()
			e.Errorf("err message")
			return t
		},
		output: `{"message":"err message","severity":"error","timestamp":"%v"}` + "\n",
	},
	{
		name:          "gcp_empty",
		cloudProvider: "GCP",
		level:         slog.LevelError,
		fields:        map[string]any{},
		logFunc: func(e Entry) time.Time {
			t := time.Now()
			e.Infof("info message")
			return t
		},
		output: ``,
	},
	{
		name:          "gcp_fields",
		cloudProvider: "GCP",
		level:         slog.LevelDebug,
		fields:        map[string]any{"field1": "value1", "field2": 2},
		logFunc: func(e Entry) time.Time {
			t := time.Now()
			e.Debugf("debug message")
			return t
		},
		output: `{"field1":"value1","field2":2,"message":"debug message","severity":"debug","timestamp":"%v"}` + "\n",
	},
	{
		name:          "gcp_logs",
		cloudProvider: "GCP",
		level:         slog.LevelDebug,
		fields:        map[string]any{},
		logFunc: func(e Entry) time.Time {
			t := time.Now()
			e.Warnf("warn message")
			return t
		},
		output: `{"message":"warn message","severity":"warning","timestamp":"%v"}` + "\n",
	},
	{
		name:          "default plain",
		format:        "plain",
		cloudProvider: "none",
		level:         slog.LevelDebug,
		fields:        map[string]any{"xyz": "abc", "abc": "xyz"},
		logFunc: func(e Entry) time.Time {
			t := time.Now()
			e.Warnf("warn message")
			return t
		},
		output: `WARNING[%v]: warn message` + "\t" + `abc=xyz xyz=abc` + " \n",
	},
	{
		name:          "default",
		cloudProvider: "none",
		level:         slog.LevelDebug,
		fields:        map[string]any{"xyz": "abc", "abc": "xyz"},
		logFunc: func(e Entry) time.Time {
			t := time.Now()
			e.Warnf("warn message")
			return t
		},
		output: `WARNING[%v]: warn message` + "\t" + `abc=xyz xyz=abc` + " \n",
	},
	{
		name:          "default json",
		format:        "json",
		cloudProvider: "none",
		level:         slog.LevelDebug,
		fields:        map[string]any{"xyz": "abc", "abc": "xyz"},
		logFunc: func(e Entry) time.Time {
			t := time.Now()
			e.Warnf("warn message")
			return t
		},
		output: `{"abc":"xyz","level":"warning","msg":"warn message","time":"%v","xyz":"abc"}` + "\n",
	},
}

func TestCloudLogger(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			lggr := New(tc.cloudProvider, tc.level, tc.format, buf)
			e := lggr.WithFields(tc.fields)
			entryTime := tc.logFunc(e)
			out := buf.String()
			expected := tc.output
			if strings.Contains(expected, "%v") {
				if tc.format == "plain" || (tc.format == "" && (tc.cloudProvider == "none" || tc.cloudProvider == "")) {
					expected = fmt.Sprintf(expected, entryTime.Format(time.Kitchen))
				} else {
					expected = fmt.Sprintf(expected, entryTime.Format(time.RFC3339))
				}
			}

			require.Equal(t, expected, out, "expected the logged entry to match the testCase output")
		})
	}
}
func TestNoOpLogger(t *testing.T) {
	l := NoOpLogger()
	require.NotPanics(t, func() { l.Info("test") })
}
