package actions

import (
	"github.com/gomods/athens/pkg/eventlog"
	"github.com/gomods/athens/pkg/eventlog/mongo"
)

// GetEventLog returns implementation of eventlog.EventLog
func GetEventLog(mongoURI string) (eventlog.Eventlog, error) {
	l, err := mongo.NewLog(mongoURI)
	return l, err
}

// NewCacheMissesLog returns impl. of eventlog.Appender
func NewCacheMissesLog(mongoURI string) (eventlog.Appender, error) {
	l, err := mongo.NewLogWithCollection(mongoURI, "cachemisseslog")
	return l, err
}
