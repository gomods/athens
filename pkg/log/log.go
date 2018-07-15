package log

import (
	"github.com/sirupsen/logrus"
)

// Logger is the main struct that any athens
// internal service should use to communicate things.
type Logger struct {
	*logrus.Logger
}

// New constructs a new logger based on the
// environment and the cloud platform it is
// running on. TODO: take cloud arg and env
// to construct the correct JSON formatter.
func New() *Logger {
	return &Logger{Logger: logrus.New()}
}

// Entry is an abstraction to the
// Logger and the logrus.Entry
// so that *Logger always creates
// an Entry copy which ensures no
// Fields are being overwritten.
type Entry interface {
	// Basic Logging Operation
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	// Attach contextual information to the logging entry
	WithFields(fields map[string]interface{}) Entry

	// SystemErr is a method that disects the error
	// and logs the appropriate level and fields for it.
	// TODO(marwan-at-work): When we have our own Error struct
	// this method will be very useful.
	SystemErr(err error)
}

// SystemErr Entry implementation.
func (l *Logger) SystemErr(err error) {
	l.Logger.Error(err)
}

// WithFields Entry implementation.
func (l *Logger) WithFields(fields map[string]interface{}) Entry {
	e := l.Logger.WithFields(fields)

	return &entry{e}
}
