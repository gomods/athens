package log

import (
	"github.com/gobuffalo/buffalo"
	"github.com/sirupsen/logrus"
)

const logEntryKey string = "log-entry-context-key"

// SetEntryInContext stores an Entry in the buffalo context
func SetEntryInContext(ctx buffalo.Context, e Entry) {
	ctx.Set(logEntryKey, e)
}

// EntryFromContext returns an Entry that has been stored in the buffalo context.
// If there is no value for the key or the type assertion fails, it returns a new
// entry from the provided logger
func EntryFromContext(ctx buffalo.Context, l *Logger) Entry {
	d := ctx.Data()
	v := d[logEntryKey]
	if v == nil {
		return &entry{Entry: logrus.NewEntry(l.Logger)}
	}

	e, ok := v.(Entry)
	if !ok {
		return &entry{Entry: logrus.NewEntry(l.Logger)}
	}

	return e
}
