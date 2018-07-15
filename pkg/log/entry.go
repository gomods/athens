package log

import (
	"github.com/sirupsen/logrus"
)

type entry struct {
	*logrus.Entry
}

func (e *entry) WithFields(fields map[string]interface{}) Entry {
	ent := e.Entry.WithFields(fields)
	return &entry{ent}
}

func (e *entry) SystemErr(err error) {
	e.Entry.Error(err)
}
