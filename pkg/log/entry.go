package log

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/gomods/athens/pkg/errors"
)

// Entry is an abstraction to the
// Logger and the slog.Entry
// so that *Logger always creates
// an Entry copy which ensures no
// Fields are being overwritten.
type Entry interface {
	// Debugf logs a debug message with formatting
	Debugf(string, ...interface{})

	// Infof logs an info message with formatting
	Infof(string, ...interface{})

	// Warnf logs a warning message with formatting
	Warnf(string, ...interface{})

	// Errorf logs an error message with formatting
	Errorf(string, ...interface{})

	// Fatalf logs a fatal message with formatting and terminates the program
	Fatalf(string, ...interface{})

	// Panicf logs a panic message with formatting and panics
	Panicf(string, ...interface{})

	// Printf logs a message with formatting at default level
	Printf(string, ...interface{})

	// Debug logs a debug message
	Debug(...interface{})

	// Info logs an info message
	Info(...interface{})

	// Warn logs a warning message
	Warn(...interface{})

	// Error logs an error message
	Error(...interface{})

	// Fatal logs a fatal message and terminates the program
	Fatal(...interface{})

	// Panic logs a panic message and panics
	Panic(...interface{})

	// Print logs a message at default level
	Print(...interface{})

	// WithFields returns a new Entry with the provided fields added
	WithFields(fields map[string]any) Entry

	// WithField returns a new Entry with a single field added
	WithField(key string, value any) Entry

	// WithError returns a new Entry with the error added to the fields
	WithError(err error) Entry

	// WithContext returns a new Entry with the context added to the fields
	WithContext(ctx context.Context) Entry

	// SystemErr handles system errors with appropriate logging levels
	SystemErr(err error)

	// WriterLevel returns an io.PipeWriter for the specified logging level
	WriterLevel(level slog.Level) *io.PipeWriter
}

type entry struct {
	logger *slog.Logger
}

func (e *entry) WithFields(fields map[string]any) Entry {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	return &entry{logger: e.logger.With(attrs...)}
}

func (e *entry) WithField(key string, value any) Entry {
	return &entry{logger: e.logger.With(key, value)}
}

func (e *entry) WithError(err error) Entry {
	return &entry{logger: e.logger.With("error", err)}
}

func (e *entry) WithContext(ctx context.Context) Entry {
	return &entry{logger: e.logger.With("context", ctx)}
}

func (e *entry) SystemErr(err error) {
	var athensErr errors.Error
	if !errors.AsErr(err, &athensErr) {
		e.Error(err.Error())
		return
	}

	ent := e.WithFields(errFields(athensErr))
	switch errors.Severity(err) {
	case slog.LevelWarn:
		ent.Warnf("%v", err)
	case slog.LevelInfo:
		ent.Infof("%v", err)
	case slog.LevelDebug:
		ent.Debugf("%v", err)
	default:
		ent.Errorf("%v", err)
	}
}

func (e *entry) Debug(args ...interface{}) {
	e.logger.Debug(fmt.Sprint(args...))
}

func (e *entry) Info(args ...interface{}) {
	e.logger.Info(fmt.Sprint(args...))
}

func (e *entry) Warn(args ...interface{}) {
	e.logger.Warn(fmt.Sprint(args...))
}

func (e *entry) Error(args ...interface{}) {
	e.logger.Error(fmt.Sprint(args...))
}

func (e *entry) Fatal(args ...interface{}) {
	e.logger.Error(fmt.Sprint(args...))
	os.Exit(1)
}

func (e *entry) Panic(args ...interface{}) {
	e.logger.Error(fmt.Sprint(args...))
}

func (e *entry) Print(args ...interface{}) {
	e.logger.Info(fmt.Sprint(args...))
}

func (e *entry) Debugf(format string, args ...interface{}) {
	e.logger.Debug(fmt.Sprintf(format, args...))
}

func (e *entry) Infof(format string, args ...interface{}) {
	e.logger.Info(fmt.Sprintf(format, args...))
}

func (e *entry) Warnf(format string, args ...interface{}) {
	e.logger.Warn(fmt.Sprintf(format, args...))
}

func (e *entry) Errorf(format string, args ...interface{}) {
	e.logger.Error(fmt.Sprintf(format, args...))
}

func (e *entry) Fatalf(format string, args ...interface{}) {
	e.logger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

func (e *entry) Panicf(format string, args ...interface{}) {
	e.logger.Error(fmt.Sprintf(format, args...)) // slog doesn't have Panic
}

func (e *entry) Printf(format string, args ...interface{}) {
	e.logger.Info(fmt.Sprintf(format, args...))
}

func (e *entry) WriterLevel(level slog.Level) *io.PipeWriter {
	reader, writer := io.Pipe()

	var logFunc func(args ...interface{})

	// Determine which log function to use based on the specified log level
	switch level {
	case slog.LevelDebug:
		logFunc = e.Debug
	case slog.LevelInfo:
		logFunc = e.Print
	case slog.LevelWarn:
		logFunc = e.Warn
	case slog.LevelError:
		logFunc = e.Error
	default:
		logFunc = e.Print
	}

	// Start a new goroutine to scan and write to logger
	go func(r *io.PipeReader, logFn func(...interface{})) {
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 65536), 65536)
		for scanner.Scan() {
			logFn(scanner.Text())
		}
		r.Close()
	}(reader, logFunc)

	return writer
}

func errFields(err errors.Error) map[string]any {
	f := map[string]any{
		"kind":    errors.KindText(err),
		"module":  err.Module,
		"version": err.Version,
		"ops":     errors.Ops(err),
	}
	return f
}
