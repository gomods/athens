package eventlog

import (
	"github.com/gomods/athens/pkg/cdn"
)

// Merger merges a single eventlog entry into the event log database. Merging
// means:
//
// - Appending the log entry into the event log
// - Modifying the module database accordingly
type Merger struct {
	getter    cdn.Getter
	deleter   cdn.Deleter
	mdSaver   cdn.MetadataSaver
	dataSaver cdn.DataSaver
	appender  Appender
}

// Merge will merge evt into the event log, ensuring that globally, all
// operations on the event log are serialized. After merging, it will take
// appropriate action on the module in storage. For example, if the event action
// is to add a new module, Merge will do the following:
//
// - Append that action to the log
// - Download module metadata and source code from the other Olympus
// - Store the module metadata & source code in its CDN
// - Add the existence of the module metadata & source code in its module metadata storage
func (m Merger) Merge(evt Event) error {
	// TODO
	return nil
}
