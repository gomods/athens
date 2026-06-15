package log

import (
	"fmt"
	"log/slog"
	"sort"

	"github.com/gomods/athens/pkg/errors"
)

// entry implements Entry on top of a *slog.Logger. The printf-style methods
// preserve the original logrus-based API so importers of pkg/log do not change.
type entry struct {
	sl *slog.Logger
}

func (e *entry) Debugf(format string, args ...any) { e.sl.Debug(fmt.Sprintf(format, args...)) }
func (e *entry) Infof(format string, args ...any)  { e.sl.Info(fmt.Sprintf(format, args...)) }
func (e *entry) Warnf(format string, args ...any)  { e.sl.Warn(fmt.Sprintf(format, args...)) }
func (e *entry) Errorf(format string, args ...any) { e.sl.Error(fmt.Sprintf(format, args...)) }

func (e *entry) WithFields(fields map[string]any) Entry {
	return &entry{sl: e.sl.With(attrsFromFields(fields)...)}
}

func (e *entry) SystemErr(err error) {
	var athensErr errors.Error
	if !errors.AsErr(err, &athensErr) {
		e.Errorf("%v", err)
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

func errFields(err errors.Error) map[string]any {
	return map[string]any{
		"operation": err.Op,
		"kind":      errors.KindText(err),
		"module":    err.Module,
		"version":   err.Version,
		"ops":       errors.Ops(err),
	}
}

// attrsFromFields converts a field map into a key-sorted slice of slog attrs so
// that log output is deterministic (logrus previously sorted keys for us).
func attrsFromFields(fields map[string]any) []any {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	attrs := make([]any, 0, len(keys))
	for _, k := range keys {
		attrs = append(attrs, slog.Any(k, fields[k]))
	}
	return attrs
}
