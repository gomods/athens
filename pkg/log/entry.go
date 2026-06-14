package log

// Entry is an abstraction to the
// Logger and the underlying slog logger
// so that *Logger always creates
// an Entry copy which ensures no
// Fields are being overwritten.
type Entry interface {
	// Basic Logging Operation
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)

	// Attach contextual information to the logging entry
	WithFields(fields map[string]any) Entry

	// SystemErr is a method that disects the error
	// and logs the appropriate level and fields for it.
	SystemErr(err error)
}
