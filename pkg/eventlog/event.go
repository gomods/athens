package eventlog

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

// Event is entry of event log specifying demand for a module.
type Event struct {
	// ID is identifier, also used as a pointer reference target.
	ID bson.ObjectId `json:"_id" bson:"_id"`
	// Time is cache-miss created/handled time.
	Time time.Time `json:"time_created" bson:"time_created"`
	// Module is module name.
	Module string `json:"module" bson:"module"`
	// Version is version of a module e.g. "1.10", "1.10-deprecated"
	Version string `json:"version" bson:"version"`
}
