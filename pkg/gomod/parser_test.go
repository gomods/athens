package parser

import (
	"bytes"
	"io"
	"testing"

	"errors"
	"github.com/stretchr/testify/assert"
)

var errReaderError = errors.New("errReader error")

type errReader struct{}

// Read implements io.Reader but returns error (not EOF)
func (e *errReader) Read(p []byte) (n int, err error) { return 0, errReaderError }

func TestParse(t *testing.T) {
	a := assert.New(t)

	var testCases = []struct {
		reader      io.Reader
		expected    string
		expectedErr error
	}{
		{bytes.NewBuffer([]byte(`module "my/thing"`)), "my/thing", nil},
		{bytes.NewBuffer([]byte(`module "github.com/gomods/athens"`)), "github.com/gomods/athens", nil},
		{bytes.NewBuffer([]byte(`module "github.com.athens/gomods"`)), "github.com.athens/gomods", nil},
		{bytes.NewBuffer([]byte(``)), "", ErrNotFound},
		{bytes.NewBuffer([]byte(`module "my/thing2`)), "", ErrNotFound},
		{bytes.NewBuffer([]byte(`module my/thing3`)), "", ErrNotFound},
		{bytes.NewBuffer([]byte(`module github.com/gomods/athens`)), "", ErrNotFound},
		{bytes.NewBuffer([]byte(`module "github.com?gomods"`)), "", ErrNotFound},
		{bytes.NewBuffer([]byte(`module "github.com.athens"`)), "", ErrNotFound},
		{bytes.NewBuffer([]byte(`module "github.com.athens"`)), "", ErrNotFound},
		{bytes.NewBuffer([]byte(`module "github.com&athens"`)), "", ErrNotFound},
		{bytes.NewBuffer([]byte(`module "?github%com&athens"`)), "", ErrNotFound},
		{bytes.NewBuffer([]byte(`foobar`)), "", ErrNotFound},
		{new(errReader), "", errReaderError},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			actual, actualErr := Parse(tc.reader)

			a.Equal(tc.expected, actual)
			a.Equal(tc.expectedErr, actualErr)
		})
	}
}
