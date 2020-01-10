package log

import (
	"bytes"
	"fmt"
	sort "sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

func getGCPFormatter() logrus.Formatter {
	return &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyTime:  "timestamp",
		},
	}
}

func getDevFormatter() logrus.Formatter {
	return devFormatter{}
}

type devFormatter struct{}

func (devFormatter) Format(e *logrus.Entry) ([]byte, error) {
	var buf bytes.Buffer
	sprint := color.CyanString
	switch e.Level {
	case logrus.DebugLevel:
		sprint = color.New(0xffccc).Sprintf
	case logrus.WarnLevel:
		sprint = color.YellowString
	case logrus.ErrorLevel:
		sprint = color.RedString
	}
	lvl := strings.ToUpper(e.Level.String())
	buf.WriteString(sprint(lvl))
	buf.WriteString("[" + e.Time.Format(time.Kitchen) + "]")
	buf.WriteString(": ")
	buf.WriteString(e.Message)
	buf.WriteByte('\t')
	for _, k := range sortFields(e.Data) {
		fmt.Fprintf(&buf, "%s=%s ", color.MagentaString(k), e.Data[k])
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

func sortFields(data logrus.Fields) []string {
	keys := []string{}
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func getDefaultFormatter() logrus.Formatter {
	return &logrus.JSONFormatter{}
}
