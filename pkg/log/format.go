package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// newHandler builds the slog.Handler for the given cloud provider and format.
// GCP gets a JSON handler whose keys/severity are remapped so Cloud Logging
// respects them; everything else uses the configured format.
func newHandler(w io.Writer, cloudProvider, format string, level slog.Level) slog.Handler {
	switch cloudProvider {
	case "GCP":
		return slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level:       level,
			ReplaceAttr: gcpReplaceAttr,
		})
	default:
		return parseFormat(w, format, level)
	}
}

func parseFormat(w io.Writer, format string, level slog.Level) slog.Handler {
	if format == "json" {
		return slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level:       level,
			ReplaceAttr: jsonReplaceAttr,
		})
	}
	return newDevHandler(w, level)
}

// gcpReplaceAttr renames slog's built-in keys to the ones Cloud Logging expects
// (timestamp/severity/message) and maps the level to a GCP LogSeverity string.
func gcpReplaceAttr(_ []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.TimeKey:
		return slog.String("timestamp", a.Value.Time().Format(time.RFC3339))
	case slog.LevelKey:
		return slog.String("severity", gcpSeverity(a.Value.Any().(slog.Level)))
	case slog.MessageKey:
		a.Key = "message"
		return a
	}
	return a
}

// jsonReplaceAttr keeps slog's native keys but normalizes the timestamp to
// RFC3339 to match what operators previously parsed.
func jsonReplaceAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.String(slog.TimeKey, a.Value.Time().Format(time.RFC3339))
	}
	return a
}

// GCP Cloud Logging severity strings. These must match the LogSeverity enum
// exactly or Cloud Logging silently falls back to DEFAULT severity, breaking
// severity-based routing and alerting.
// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
const (
	gcpSeverityDebug   = "DEBUG"
	gcpSeverityInfo    = "INFO"
	gcpSeverityWarning = "WARNING"
	gcpSeverityError   = "ERROR"
)

// gcpSeverity maps an slog.Level onto a canonical GCP LogSeverity string.
func gcpSeverity(l slog.Level) string {
	switch {
	case l <= slog.LevelDebug:
		return gcpSeverityDebug
	case l < slog.LevelWarn:
		return gcpSeverityInfo
	case l < slog.LevelError:
		return gcpSeverityWarning
	default:
		return gcpSeverityError
	}
}

const lightGrey = 0xffccc

// devHandler is a human-friendly, colored slog.Handler used for local
// development. It renders "LEVEL[3:04PM]: message\tkey=value " lines with
// fields sorted for deterministic output.
type devHandler struct {
	mu    *sync.Mutex
	w     io.Writer
	level slog.Level
	attrs []slog.Attr
}

func newDevHandler(w io.Writer, level slog.Level) *devHandler {
	return &devHandler{mu: &sync.Mutex{}, w: w, level: level}
}

func (h *devHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level
}

func (h *devHandler) Handle(_ context.Context, r slog.Record) error {
	var sprintf func(format string, a ...any) string
	switch {
	case r.Level <= slog.LevelDebug:
		sprintf = color.New(lightGrey).Sprintf
	case r.Level == slog.LevelWarn:
		sprintf = color.YellowString
	case r.Level >= slog.LevelError:
		sprintf = color.RedString
	default:
		sprintf = color.CyanString
	}

	var buf bytes.Buffer
	buf.WriteString(sprintf(strings.ToUpper(r.Level.String())))
	buf.WriteString("[" + r.Time.Format(time.Kitchen) + "]")
	buf.WriteString(": ")
	buf.WriteString(r.Message)
	buf.WriteByte('\t')

	attrs := make([]slog.Attr, 0, len(h.attrs)+r.NumAttrs())
	attrs = append(attrs, h.attrs...)
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})
	sort.Slice(attrs, func(i, j int) bool { return attrs[i].Key < attrs[j].Key })
	for _, a := range attrs {
		fmt.Fprintf(&buf, "%s=%s ", color.MagentaString(a.Key), a.Value)
	}
	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf.Bytes())
	return err
}

func (h *devHandler) WithAttrs(as []slog.Attr) slog.Handler {
	nh := *h
	nh.attrs = make([]slog.Attr, 0, len(h.attrs)+len(as))
	nh.attrs = append(nh.attrs, h.attrs...)
	nh.attrs = append(nh.attrs, as...)
	return &nh
}

func (h *devHandler) WithGroup(string) slog.Handler { return h }
