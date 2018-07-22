package errors

import (
	"errors"
	"net/http"
	"strings"

	"github.com/marwan-at-work/vgop/semver"
	"github.com/sirupsen/logrus"
)

// Kind enums
const (
	KindNotFound   = http.StatusNotFound
	KindBadRequest = http.StatusBadRequest
	KindUnexpected = http.StatusInternalServerError
)

// Error is an Athens system error.
// It carries information and behavior
// as to what caused this error so that
// callers can implement logic around it.
type Error struct {
	// Kind categories Athens errors into a smaller
	// subset of errors. This way we can generalize
	// what an error really is: such as "not found",
	// "bad request", etc. The official categories
	// are HTTP status code but the ones we use are
	// imported into this package.
	Kind     int
	Op       Op
	Module   string
	Version  string
	Err      error
	Severity logrus.Level
}

// Error returns the underlying error's
// string message. The logger takes care
// of filling out the stack levels and
// extra information.
func (e Error) Error() string {
	return e.Err.Error()
}

// Op describes any independent function or
// method in Athens. A series of operations
// forms a more readable stack trace.
type Op string

// E is a helper function to construct an Error type
// Operation always comes first, module path and version
// come second, they are optional. Args must have at least
// an error or a string to describe what exactly went wrong.
// You can optionally pass a Logrus severity to indicate
// the log level of an error based on the context it was constructed in.
func E(op Op, args ...interface{}) error {
	if len(args) == 0 {
		panic("errors.E called with 0 args")
	}
	e := Error{Op: op}
	for _, a := range args {
		switch a := a.(type) {
		case error:
			e.Err = a
		case string:
			switch {
			case isModule(a):
				e.Module = a
			case isVersion(a):
				e.Version = a
			default:
				e.Err = errors.New(a)
			}
		case logrus.Level:
			e.Severity = a
		case int:
			e.Kind = a
		}
	}

	if e.Err == nil {
		e.Err = errors.New("no error message provided")
	}

	return e
}

// Severity returns the log level of an error
// if non exists, then it's an Error level because
// it is an unexpected.
func Severity(err error) logrus.Level {
	e, ok := err.(Error)
	if !ok {
		return logrus.ErrorLevel
	}

	// if there's no severity (0 is Panic level in logrus
	// which we should not use since cloud providers only have
	// debug, info, warn, and error) then look for the
	// child's severity.
	if childE, ok := e.Err.(Error); ok && e.Severity == 0 {
		return Severity(childE)
	}

	return e.Severity
}

// Kind recursively searches for the
// first error kind it finds.
func Kind(err error) int {
	e, ok := err.(Error)
	if !ok {
		return KindUnexpected
	}

	if e.Kind != 0 {
		return e.Kind
	}

	return Kind(e.Err)
}

// KindText returns a friendly string
// of the Kind type. Since we use http
// status codes to represent error kinds,
// this method just deferrs to the net/http
// text representations of statuses.
func KindText(err error) string {
	return http.StatusText(Kind(err))
}

// Ops aggregates the error's operation
// with all the embedded errors' operations.
// This way you can construct a queryable
// stack trace.
func Ops(err Error) []Op {
	ops := []Op{err.Op}
	for {
		embeddedErr, ok := err.Err.(Error)
		if !ok {
			break
		}

		ops = append(ops, embeddedErr.Op)
	}

	return ops
}

func isModule(s string) bool {
	ss := strings.Split(s, "/")
	return len(ss) > 1 && strings.Contains(ss[0], ".")
}

func isVersion(s string) bool {
	return semver.IsValid(s)
}
