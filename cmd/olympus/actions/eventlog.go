package actions

import (
	"github.com/gomods/athens/pkg/env"
	"github.com/gomods/athens/pkg/eventlog"
	"github.com/gomods/athens/pkg/eventlog/mongo"
)

// GetEventLog returns implementation of eventlog.EventLog
func GetEventLog() (eventlog.Eventlog, error) {
	deets, err := env.ForMongo()
	if err != nil {
		return nil, err
	}
	l, err := mongo.NewLog(deets, mongo.EventLogCollection)
	return l, err
}

func newCacheMissesLog() (eventlog.Appender, error) {
	deets, err := env.ForMongo()
	if err != nil {
		return nil, err
	}
	return mongo.NewLog(deets, mongo.CacheMissLogCollection)
}
