package eventlog

import (
	"github.com/gomods/athens/pkg/storage"
)

// Merger merges a single eventlog entry into the event log database. Merging
// means:
//
// - Appending the log entry into the event log
// - Modifying the module database accordingly
type Merger struct {
	getter   storage.Getter
	deleter  storage.Deleter
	setter   storage.Saver
	appender Appender
}

// Merge will merge evt into the event log, ensuring that globally, all
// operations on the event log are serialized. After merging, it will take
// appropriate action on the module in storage. For example, if the event action
// is to add a new module, Merge will append that action to the log and add
// the module to storage.
//
// If the receiving Olympus (OX) gets an add event from OY, OX will add the
// module in a two step process:
//
// - Store the CDN location as OY's CDN
// - Download module metadata and source from OY's CDN
// - Update the CDN location to OX's CDN
func (m Merger) Merge(evt Event) error {
	// TODO
	return nil
}
