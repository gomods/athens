package log

import (
	"bytes"
	"fmt"
	"sort"
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

const lightGrey = 0xffccc

func (devFormatter) Format(e *logrus.Entry) ([]byte, error) {
	var buf bytes.Buffer
	var sprintf func(format string, a ...any) string
	switch e.Level {
	case logrus.DebugLevel:
		sprintf = color.New(lightGrey).Sprintf
	case logrus.WarnLevel:
		sprintf = color.YellowString
	case logrus.ErrorLevel:
		sprintf = color.RedString
	default:
		sprintf = color.CyanString
	}
	lvl := strings.ToUpper(e.Level.String())
	buf.WriteString(sprintf(lvl))
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

func parseFormat(format string) logrus.Formatter {
	if format == "json" {
		return &logrus.JSONFormatter{}
	}

	return getDevFormatter()
}
