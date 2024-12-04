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

type Entry interface {
	// Keep the existing interface methods unchanged
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
	Panicf(string, ...interface{})
	Printf(string, ...interface{})

	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	Fatal(...interface{})
	Panic(...interface{})
	Print(...interface{})

	WithFields(fields map[string]any) Entry
	WithField(key string, value any) Entry
	WithError(err error) Entry
	WithContext(ctx context.Context) Entry
	SystemErr(err error)
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
	e.logger.Error(fmt.Sprint(args...)) // slog doesn't have Fatal, using Error
}

func (e *entry) Panic(args ...interface{}) {
	e.logger.Error(fmt.Sprint(args...)) // slog doesn't have Panic, using Error
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
