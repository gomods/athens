package actions

import (
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/eventlog"
	"github.com/gomods/athens/pkg/eventlog/mongo"
)

// GetEventLog returns implementation of eventlog.EventLog
func GetEventLog(mongoURL string, certPath string) (eventlog.Eventlog, error) {
	const op = "actions.GetEventLog"
	l, err := mongo.NewLog(mongoURL, certPath)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return l, nil
}

// NewCacheMissesLog returns impl. of eventlog.Appender
func NewCacheMissesLog(mongoURL string, certPath string) (eventlog.Appender, error) {
	const op = "actions.NewCacheMissesLog"
	l, err := mongo.NewLogWithCollection(mongoURL, certPath, "cachemisseslog")
	if err != nil {
		return nil, errors.E(op, err)
	}
	return l, nil
}
