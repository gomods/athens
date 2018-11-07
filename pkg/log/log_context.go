package log

import (
	"github.com/gobuffalo/buffalo"
)

const logEntryKey = "log-entry-context-key"

// SetEntryInContext stores an Entry in the buffalo context
func SetEntryInContext(ctx buffalo.Context, e Entry) {
	ctx.Set(logEntryKey, e)
}

// EntryFromContext returns an Entry that has been stored in the buffalo context.
// If there is no value for the key or the type assertion fails, it returns a new
// entry from the provided logger
func EntryFromContext(ctx buffalo.Context) Entry {
	d := ctx.Data()
	e, ok := d[logEntryKey].(Entry)
	if e == nil || !ok {
		return NoOpLogger()
	}

	return e
}
