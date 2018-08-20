package env

import (
	"github.com/gobuffalo/envy"
	"github.com/sirupsen/logrus"
)

// LogLevel returns the system's
// exposure to internal logs. Defaults
// to debug.
func LogLevel() (logrus.Level, error) {
	lvlStr := envy.Get("ATHENS_LOG_LEVEL", "debug")
	return logrus.ParseLevel(lvlStr)
}
