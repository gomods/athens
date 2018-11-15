package log

import (
	"time"

	"github.com/sirupsen/logrus"
)

// DefaultTimeStampFormat is the default timestamp format for the logger
// RFC3339     = "2006-01-02T15:04:05Z07:00"
const DefaultTimeStampFormat string = time.RFC3339

// Logger is the main struct that any athens
// internal service should use to communicate things.
type Logger struct {
	*logrus.Logger
}

// New constructs a new logger based on the
// environment and the cloud platform it is
// running on. TODO: take cloud arg and env
// to construct the correct JSON formatter.
func New(cloudProvider string, level logrus.Level, timestampFormat string) *Logger {
	l := logrus.New()
	switch cloudProvider {
	case "GCP":
		l.Formatter = getGCPFormatter(timestampFormat)
	case "none":
		l.Formatter = getDevFormatter(timestampFormat)
	default:
		l.Formatter = getDefaultFormatter(timestampFormat)
	}
	l.Level = level
	return &Logger{Logger: l}
}

// SystemErr Entry implementation.
func (l *Logger) SystemErr(err error) {
	e := &entry{Entry: logrus.NewEntry(l.Logger)}
	e.SystemErr(err)
}

// WithFields Entry implementation.
func (l *Logger) WithFields(fields map[string]interface{}) Entry {
	e := l.Logger.WithFields(fields)

	return &entry{e}
}

// NoOpLogger provides a Logger that does nothing
func NoOpLogger() *Logger {
	return &Logger{
		Logger: &logrus.Logger{},
	}
}
