// Package build provides details of the built binary
// The details are set using ldflags.
//
// The ldflags can be set either manually:
// `go build -ldflags "-X github.com/gomods/athens/pkg/build.commitSHA=$(git rev-list -1 HEAD) -X github.com/gomods/athens/pkg/build.version=$(git describe --tags) -X github.com/gomods/athens/pkg/build.buildDate$(date -u +%Y-%m-%d-%H:%M:%S-%Z)"`
//
// or using the build script in ./scripts.
package build

import "fmt"

var commitSHA, version, buildDate string

// InfoString returns build details as a string with formatting
// suitable for console output.
func InfoString() string {
	return fmt.Sprintf("Build Details:\n\tVersion:\t%s\n\tCommit SHA:\t%s\n\tDate:\t\t%s", version, commitSHA, buildDate)
}

// JSON returns a JSON thing for use in server response bodies
func JSON() string {
	// TODO
	return ""
}

// Commit returns the build's commit hash
func Commit() string {
	return commitSHA
}

// Version returns the build's version string
func Version() string {
	return version
}

// Date returns the formatted date the binary was built
func Date() string {
	return buildDate
}
