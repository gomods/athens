package log

import (
	"github.com/sirupsen/logrus"
)

func getGCPFormatter(timestampFormat string) logrus.Formatter {
	ft := &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyTime:  "timestamp",
		},
	}
	if timestampFormat != "" {
		ft.TimestampFormat = timestampFormat
	}
	return ft
}

func getDevFormatter(timestampFormat string) logrus.Formatter {
	ft := &logrus.TextFormatter{}
	if timestampFormat != "" {
		ft.TimestampFormat = timestampFormat
	}
	return ft
}

func getDefaultFormatter(timestampFormat string) logrus.Formatter {
	ft := &logrus.JSONFormatter{}
	if timestampFormat != "" {
		ft.TimestampFormat = timestampFormat
	}
	return ft
}
