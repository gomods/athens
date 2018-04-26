package eventlog

// Eventlog is append only log of Events.
type Eventlog interface {
	Reader
	Appender
}

// Reader is reader of append only event log.
type Reader interface {
	// Read reads all events in event log.
	Read() []Event
	// ReadFrom reads all events from the log starting at event with specified id (excluded).
	// If id is not found behaves like Read().
	ReadFrom(id int64) []Event
}

// Appender is writer to append only event log.
type Appender interface {
	// Write appends Event to event log and returns its ID.
	Write(event Event) int64
}
