package log

import (
	"context"
	"log/slog"

	"github.com/gomods/athens/pkg/errors"
)

// Entry is an abstraction to the
// Logger and the logrus.Entry
// so that *Logger always creates
// an Entry copy which ensures no
// Fields are being overwritten.
type Entry interface {
	// Basic Logging Operation
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)

	// Attach contextual information to the logging entry
	WithFields(fields map[string]any) Entry

	WithField(key string, value any) Entry

	WithError(err error) Entry

	WithContext(ctx context.Context) Entry

	// SystemErr is a method that disects the error
	// and logs the appropriate level and fields for it.
	SystemErr(err error)
}

type entry struct {
	*slog.Logger
}

func (e *entry) WithFields(fields map[string]any) Entry {
	ent := e.WithFields(fields)
	return ent
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
func errFields(err errors.Error) map[string]any {
	f := map[string]any{}
	f["operation"] = err.Op
	f["kind"] = errors.KindText(err)
	f["module"] = err.Module
	f["version"] = err.Version
	f["ops"] = errors.Ops(err)

	return f
}
