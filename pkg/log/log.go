package log

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
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
func New(cloudProvider string, level slog.Level, format string, w io.Writer) *Logger {
	var l *slog.Logger
	switch cloudProvider {
	case "GCP":
		l = getGCPFormatter(level, w)
	default:
		l = parseFormat(format, level, w)
	}
	if l == nil {
		l = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))
	}
	slog.SetDefault(l)
	return &Logger{Logger: l}
}

// SystemErr Entry implementation.
func (l *Logger) SystemErr(err error) {
	e := &entry{l.Logger}
	e.SystemErr(err)
}

// WithFields Entry implementation.
func (l *Logger) WithFields(fields map[string]any) Entry {
	attrs := make([]any, 0, len(fields))
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	return &entry{logger: l.Logger.With(attrs...)}
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

func (l *Logger) WriterLevel(level slog.Level) *io.PipeWriter {
	pipeReader, pipeWriter := io.Pipe()
	go func() {
		scanner := bufio.NewScanner(pipeReader)
		for scanner.
			Scan() {
			l.Info(scanner.Text())
		}
	}()
	return pipeWriter
}

func (l *Logger) Fatal(args ...any) {
	l.Logger.Error(fmt.Sprint(args...))
	os.Exit(1)
}

// NoOpLogger provides a Logger that does nothing.
func NoOpLogger() *Logger {
	return &Logger{
		Logger: &slog.Logger{},
	}
}
