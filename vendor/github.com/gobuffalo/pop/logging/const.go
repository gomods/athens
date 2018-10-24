package logging

// Level is the logger level
type Level int

const (
	// SQL level is the lowest logger level. It dumps all logs.
	SQL Level = iota
	// Debug level dumps logs with higher or equal severity than debug.
	Debug
	// Info level dumps logs with higher or equal severity than info.
	Info
	// Warn level dumps logs with higher or equal severity than warning.
	Warn
	// Error level dumps logs only errors.
	Error
)

func (l Level) String() string {
	switch l {
	case SQL:
		return "sql"
	case Debug:
		return "debug"
	case Info:
		return "info"
	case Warn:
		return "warn"
	case Error:
		return "error"
	}
	return "unknown"
}
