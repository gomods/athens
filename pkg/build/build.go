// Package build provides details of the built binary
// The details are set using ldflags.
//
// The ldflags can be set either manually:
// `go build -ldflags "-X github.com/gomods/athens/pkg/build.commitSHA=$(git rev-list -1 HEAD) -X github.com/gomods/athens/pkg/build.version=$(git describe --tags) -X github.com/gomods/athens/pkg/build.buildDate$(date -u +%Y-%m-%d-%H:%M:%S-%Z)"`
//
// or using the build script in ./scripts.
package build

import (
	"encoding/json"
	"fmt"
)

// details represents known data for a given build
type details struct {
	Version string
	Commit  string
	Date    string
}

var commitSHA, version, buildDate string

// String returns build details as a string with formatting
// suitable for console output.
//
// i.e.
// Build Details:
//         Version:        v0.1.0-155-g1a20f8b
//         Commit SHA:     1a20f8b6a36136183f8533ae850a582716bbd577
//         Date:           2018-11-05-14:33:14-UTC
func String() string {
	return fmt.Sprintf("Build Details:\n\tVersion:\t%s\n\tCommit SHA:\t%s\n\tDate:\t\t%s", version, commitSHA, buildDate)
}

// JSON returns build details JSON object as a slice of byte
// for use in server response bodies
func JSON() ([]byte, error) {
	out := details{
		Version: version,
		Commit:  commitSHA,
		Date:    buildDate,
	}
	return json.Marshal(out)
}
