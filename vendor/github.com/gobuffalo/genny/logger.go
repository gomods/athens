package genny

// Logger interface for a logger to be used
// with genny. Logrus is 100% compatible.
type Logger interface {
	Debugf(string, ...interface{})
	Debug(...interface{})
	Infof(string, ...interface{})
	Info(...interface{})
	Printf(string, ...interface{})
	Print(...interface{})
	Warnf(string, ...interface{})
	Warn(...interface{})
	Errorf(string, ...interface{})
	Error(...interface{})
	Fatalf(string, ...interface{})
	Fatal(...interface{})
}
