package storage

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHistoricalReleaseFileIdentity(t *testing.T) {
	for _, file := range []ReleaseFile{
		{Filename: "go1.22.12.windows-arm.zip", Version: "go1.22.12", OS: "windows", Arch: "armv6l", Kind: "archive"},
		{Filename: "go1.22.12.linux-armv6l.tar.gz", Version: "go1.22.12", OS: "linux", Arch: "armv6l", Kind: "archive"},
		{Filename: "go1.4-bootstrap-20170518.tar.gz", Version: "go1", OS: "4", Arch: "bootstrap", Kind: "archive"},
	} {
		file.SHA256 = strings.Repeat("0", 64)
		file.Size = 1
		require.NoError(t, file.Validate(), file.Filename)
		file.OS = "wrong"
		require.Error(t, file.Validate(), file.Filename)
	}
}
