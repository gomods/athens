package log

import (
	"fmt"
	"io"
	stdlog "log"
	"log/slog"
	"os"
	"strings"
)

// Logger is the main struct that any athens
// internal service should use to communicate things.
type Logger struct {
	*entry

	handler slog.Handler
	level   slog.Level
}

// New constructs a new logger based on the
// environment and the cloud platform it is
// running on.
func New(cloudProvider string, level slog.Level, format string) *Logger {
	return NewWithOutput(os.Stdout, cloudProvider, level, format)
}

// NewWithOutput is New but writes to the given io.Writer instead of stdout.
// It is primarily useful for tests that need to capture log output.
func NewWithOutput(w io.Writer, cloudProvider string, level slog.Level, format string) *Logger {
	h := newHandler(w, cloudProvider, format, level)
	return &Logger{
		entry:   &entry{sl: slog.New(h)},
		handler: h,
		level:   level,
	}
}

// Fatalf logs at the error level and then exits the process. slog has no
// fatal level, so this preserves logrus.Fatal's exit-after-logging behavior.
func (l *Logger) Fatalf(format string, args ...any) {
	l.Errorf(format, args...)
	os.Exit(1)
}

// StdLogger returns a *log.Logger that routes through this logger's handler at
// the given level. Used to redirect output from the standard library logger.
func (l *Logger) StdLogger(level slog.Level) *stdlog.Logger {
	return slog.NewLogLogger(l.handler, level)
}

// NoOpLogger provides a Logger that does nothing.
func NoOpLogger() *Logger {
	return &Logger{
		entry:   &entry{sl: slog.New(slog.DiscardHandler)},
		handler: slog.DiscardHandler,
	}
}

// ParseLevel converts a configured log-level string into an slog.Level. It
// accepts the legacy logrus level names so existing operator configuration
// keeps working: trace maps to debug, warning to warn, and fatal/panic to error.
func ParseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "trace", "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error", "fatal", "panic":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("not a valid log level: %q", s)
	}
}
