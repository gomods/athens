package log

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gobuffalo/buffalo"
	"github.com/sirupsen/logrus"
)

// Buffalo returns a more sane logging format
// than the default buffalo formatter.
// For the most part, we only care about
// the path, the method, and the status code.
// It's also good to note that internal logs
// from buffalo should only be allowed in development
// as our logging-system should be handled from our codebase.
func Buffalo() buffalo.Logger {
	l := logrus.New()
	l.Formatter = &buffaloFormatter{}

	return &buffaloLogger{l}
}

type buffaloFormatter struct{}

func (buffaloFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if entry.Level == logrus.ErrorLevel {
		// buffalo does not pass request params when an error occurs: pass params
		// when https://github.com/gobuffalo/buffalo/issues/1171 is resolved.
		return fmtBuffaloErr(entry.Message), nil
	}

	statusCode, _ := entry.Data["status"].(int)
	status := fmt.Sprint(statusCode)

	switch {
	case statusCode < 400:
		status = color.GreenString("%v", status)
	case statusCode >= 400 && statusCode < 500:
		status = color.HiYellowString("%v", status)
	default:
		status = color.HiRedString("%v", status)
	}

	str := fmt.Sprintf(
		"%v %v %v [%v]\n",
		color.CyanString("handler:"),
		entry.Data["method"],
		entry.Data["path"],
		status,
	)

	return []byte(str), nil
}

type buffaloLogger struct{ logrus.FieldLogger }

func (bf *buffaloLogger) WithField(key string, val interface{}) buffalo.Logger {
	e := bf.FieldLogger.WithField(key, val)

	return &buffaloLogger{e}
}

func (bf *buffaloLogger) WithFields(fields map[string]interface{}) buffalo.Logger {
	e := bf.FieldLogger.WithFields(fields)
	return &buffaloLogger{e}
}

func fmtBuffaloErr(msg string) []byte {
	return []byte(fmt.Sprintf("%s %s\n", color.HiRedString("buffalo:"), msg))
}
