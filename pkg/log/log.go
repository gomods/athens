package log

import (
	"bytes"
	"context"
	"log/slog"
	"os"
)

// Logger is the main struct that any athens
// internal service should use to communicate things.
type Logger struct {
	*slog.Logger

	Out *bytes.Buffer
}

// New constructs a new logger based on the
// environment and the cloud platform it is
// running on. TODO: take cloud arg and env
// to construct the correct JSON formatter.
func New(cloudProvider string, level slog.Level, format string) *Logger {
	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	switch cloudProvider {
	case "GCP":
		l = getGCPFormatter(level)
	default:
		l = parseFormat(format, level)
	}
	slog.SetDefault(l)
	return &Logger{Logger: l}
}

// SystemErr Entry implementation.
func (l *Logger) SystemErr(err error) {
	e := &entry{Logger: l.Logger}
	e.SystemErr(err)
}

// WithFields Entry implementation.
func (l *Logger) WithFields(fields map[string]any) Entry {
	return l.WithFields(fields)
}

func (l *Logger) WithField(key string, value any) Entry {
	keys := map[string]any{
		key: value,
	}
	return l.WithFields(keys)
}

func (l *Logger) WithError(err error) Entry {
	keys := map[string]any{
		"error": err,
	}
	return l.WithFields(keys)
}

func (l *Logger) WithContext(ctx context.Context) Entry {
	keys := map[string]any{
		"context": ctx,
	}
	return l.WithFields(keys)
}

// NoOpLogger provides a Logger that does nothing.
func NoOpLogger() *Logger {
	l := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return &Logger{Logger: l}
}
