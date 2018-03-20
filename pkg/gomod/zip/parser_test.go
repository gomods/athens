package zip

import (
	"archive/zip"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gomods/athens/pkg/gomod"
)

func TestZipParser_ModuleName(t *testing.T) {
	a := assert.New(t)

	var testCases = []struct {
		zipfile     string
		expected    string
		expectedErr error
	}{
		{"testdata/go.0.zip", "my/thing", nil},
		{"testdata/go.1.zip", "my/thing2", nil},
		{"testdata/go.2.zip", "", parser.ErrNotFound},
		{"testdata/go.3.zip", "", parser.ErrNotFound},
		{"testdata/go.4.zip", "", errors.New("go.mod not found")},
	}

	for _, tc := range testCases {
		t.Run(tc.zipfile, func(t *testing.T) {
			reader, err := zip.OpenReader(tc.zipfile)
			a.NoError(err)
			defer reader.Close()
			fp := NewZipParser(*reader)
			actual, actualErr := fp.ModuleName()

			a.Equal(tc.expected, actual)
			a.Equal(tc.expectedErr, actualErr)
		})
	}
}
