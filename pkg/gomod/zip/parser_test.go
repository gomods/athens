package zip

import (
	"archive/zip"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gomods/athens/pkg/gomod"
	"io/ioutil"
	"io"
	"os"
	"github.com/stretchr/testify/require"
)

func TestZipParser_ModuleName(t *testing.T) {
	a := assert.New(t)

	var testCases = []struct {
		zipfile     string
		expected    string
		expectedErr error
	}{
		{zipTestMod(t, "testdata/go.0.mod", "go.mod"), "my/thing", nil},
		{zipTestMod(t, "testdata/go.1.mod", "go.mod"), "my/thing2", nil},
		{zipTestMod(t, "testdata/go.2.mod", "go.mod"), "", parser.ErrNotFound},
		{zipTestMod(t, "testdata/go.3.mod", "go.mod"), "", parser.ErrNotFound},
		{zipTestMod(t, "testdata/go.4.mod", "go.mod"), "", parser.ErrNotFound},
		{zipTestMod(t, "testdata/Gopkg.toml", "Gopkg.toml"), "", errors.New("go.mod not found")},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
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

func zipTestMod(t *testing.T, src string, dstFileName string) (target string) {
	r := require.New(t)
	zipfile, err := ioutil.TempFile("", "")
	r.NoError(err, "an error occurred while creating temporary file")
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	srcfile, err := os.Open(src)
	r.NoError(err, "an error occurred while opening fixture file")
	defer srcfile.Close()

	f, err := archive.Create(dstFileName)
	r.NoError(err, "an error occurred while creating file in archive")

	_, err = io.Copy(f, srcfile)
	r.NoError(err, "an error occurred while coping data to archive")

	return zipfile.Name()
}