package log

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

type input struct {
	name          string
	cloudProvider string
	level         string
	fields        logrus.Fields
	logFunc       func(e Entry)
	output        string
}

var testCases = []input{
	{
		"gcp_debug",
		"GCP",
		"debug",
		logrus.Fields{},
		func(e Entry) { e.Infof("info message") },
		`{"message":"info message","severity":"info","timestamp":"%v"}` + "\n",
	},
	{
		"gcp_error",
		"GCP",
		"debug",
		logrus.Fields{},
		func(e Entry) { e.Errorf("err message") },
		`{"message":"err message","severity":"error","timestamp":"%v"}` + "\n",
	},
	{
		"gcp_empty",
		"GCP",
		"error",
		logrus.Fields{},
		func(e Entry) { e.Infof("info message") },
		``,
	},
	{
		"gcp_fields",
		"GCP",
		"debug",
		logrus.Fields{"field1": "value1", "field2": 2},
		func(e Entry) { e.Debugf("debug message") },
		`{"field1":"value1","field2":2,"message":"debug message","severity":"debug","timestamp":"%v"}` + "\n",
	},
	{
		"gcp_logs",
		"GCP",
		"debug",
		logrus.Fields{},
		func(e Entry) { e.Warnf("warn message") },
		`{"message":"warn message","severity":"warning","timestamp":"%v"}` + "\n",
	},
	{
		"default",
		"default",
		"debug",
		logrus.Fields{"xyz": "abc", "abc": "xyz"},
		func(e Entry) { e.Warnf("warn message") },
		`{"abc":"xyz","level":"warning","msg":"warn message","time":"%v","xyz":"abc"}` + "\n",
	},
}

func TestCloudLogger(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lggr := New(tc.cloudProvider, tc.level)
			var buf bytes.Buffer
			lggr.Out = &buf
			e := lggr.WithFields(tc.fields)
			tc.logFunc(e)
			out := buf.String()
			expected := tc.output
			if strings.Contains(expected, "%v") {
				expected = fmt.Sprintf(tc.output, time.Now().Format(time.RFC3339))
			}

			if expected != out {
				t.Fatalf("expected to log %q but got %q", expected, out)
			}
		})
	}
}
