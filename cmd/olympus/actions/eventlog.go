package actions

import (
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/eventlog"
	"github.com/gomods/athens/pkg/eventlog/mongo"
)

// GetEventLog returns implementation of eventlog.EventLog
func GetEventLog(mongoURL string, certPath string, timeout time.Duration) (eventlog.Eventlog, error) {
	const op = "actions.GetEventLog"
	l, err := mongo.NewLog(mongoURL, certPath, timeout)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return l, nil
}

// NewCacheMissesLog returns impl. of eventlog.Appender
func NewCacheMissesLog(mongoURL string, certPath string, timeout time.Duration) (eventlog.Appender, error) {
	const op = "actions.NewCacheMissesLog"
	l, err := mongo.NewLogWithCollection(mongoURL, certPath, "cachemisseslog", timeout)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return l, nil
}
