package formam

import (
	"encoding/json"
)

type Error struct {
	err error
}

func (s *Error) Error() string {
	return "formam: " + s.err.Error()
}

func (s Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.err.Error())
}

// Cause implements the causer interface from github.com/pkg/errors.
func (s *Error) Cause() error {
	return s.err
}

func newError(err error) *Error { return &Error{err} }
