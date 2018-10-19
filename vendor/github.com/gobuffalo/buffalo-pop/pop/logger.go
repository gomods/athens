package pop

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/pop/logging"
)

type logger = func(lvl logging.Level, s string, args ...interface{})

func Logger(app *buffalo.App) logger {
	return func(lvl logging.Level, s string, args ...interface{}) {
		if !pop.Debug && lvl <= logging.Debug {
			return
		}
		l := app.Logger
		if pop.Color {
			s = color.YellowString(s)
		}
		switch lvl {
		case logging.SQL:
			if len(args) > 0 {
				for i, a := range args {
					l = l.WithField(fmt.Sprintf("$%d", i+1), a)
				}
			}
			l.Debug(s)
		case logging.Debug:
			l.Debugf(s, args...)
		case logging.Info:
			l.Infof(s, args...)
		case logging.Warn:
			l.Warnf(s, args...)
		default:
			l.Printf(s, args...)
		}
	}
}
