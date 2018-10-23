package pop

import (
	"fmt"
	stdlog "log"
	"os"

	"github.com/fatih/color"
	"github.com/gobuffalo/pop/logging"
	"github.com/markbates/oncer"
)

type logger func(lvl logging.Level, s string, args ...interface{})
type legacyLogger func(s string, args ...interface{})

// Debug mode, to toggle verbose log traces
var Debug = false

// Color mode, to toggle colored logs
var Color = true

var log logger

var defaultStdLogger = stdlog.New(os.Stdout, "[POP] ", stdlog.LstdFlags)
var defaultLogger = func(lvl logging.Level, s string, args ...interface{}) {
	// Handle legacy logger
	if Log != nil {
		oncer.Deprecate(0, "pop.Log", "Use pop.SetLogger instead.")
		Log(s, args...)
		return
	}
	if !Debug && lvl <= logging.Debug {
		return
	}
	if lvl == logging.SQL {
		if len(args) > 0 {
			xargs := make([]string, len(args))
			for i, a := range args {
				switch a.(type) {
				case string:
					xargs[i] = fmt.Sprintf("%q", a)
				default:
					xargs[i] = fmt.Sprintf("%v", a)
				}
			}
			s = fmt.Sprintf("%s - %s | %s", lvl, s, xargs)
		} else {
			s = fmt.Sprintf("%s - %s", lvl, s)
		}
	} else {
		s = fmt.Sprintf(s, args...)
		s = fmt.Sprintf("%s - %s", lvl, s)
	}
	if Color {
		s = color.YellowString(s)
	}
	defaultStdLogger.Println(s)
}

// SetLogger overrides the default logger.
//
// The logger must implement the following interface:
// type logger func(lvl logging.Level, s string, args ...interface{})
func SetLogger(l logger) {
	log = l
}

// Log defines the pop logger. Override it to customize pop logs handling.
// Deprecated: use SetLogger instead
var Log legacyLogger
