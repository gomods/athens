package log

import (
	"github.com/sirupsen/logrus"
)

func getGCPFormatter(timestampFormat string) logrus.Formatter {
	return &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyTime:  "timestamp",
		},
		TimestampFormat: timestampFormat,
	}
}

func getDevFormatter(timestampFormat string) logrus.Formatter {
	return &logrus.TextFormatter{TimestampFormat: timestampFormat}
}

func getDefaultFormatter(timestampFormat string) logrus.Formatter {
	return &logrus.JSONFormatter{TimestampFormat: timestampFormat}
}
