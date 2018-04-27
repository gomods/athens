package eventlog

// Eventlog is append only log of Events.
type Eventlog interface {
	Reader
	Appender
	Clearer
}

// Reader is reader of append only event log.s
type Reader interface {
	// Read reads all events in event log.
	Read() ([]Event, error)

	// ReadFrom reads all events from the log starting at event with specified id (excluded).
	// If id is not found behaves like Read().
	ReadFrom(id string) ([]Event, error)
}

// Appender is writer to append only event log.
type Appender interface {
	// Write appends Event to event log and returns its ID.
	Write(event Event) (string, error)
}

// Clearer is interface used to clear state of event log
type Clearer interface {
	Clear(id string) error
}
