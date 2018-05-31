package eventlog

// Merger merges a single eventlog entry into the event log database. Merging
// means:
//
// - Appending the log entry into the event log
// - Modifying the module database accordingly
type Merger struct {
	getter   storage.Getter
	appender Appender
}

func (m Merger) Merge(evt Event) error {
	// TODO
}
