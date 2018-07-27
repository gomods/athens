package env

import (
	"fmt"

	"github.com/gobuffalo/envy"
)

// GoEnvironmentWithDefault returns environment used.
// Supported values are: 'development' and 'production'
func GoEnvironmentWithDefault(value string) string {
	return envy.Get("GO_ENV", value)
}

// GoPath returns gopath defined by env variable GOPATH.
func GoPath() (string, error) {
	env, err := envy.MustGet("GOPATH")
	if err != nil {
		return "", fmt.Errorf("GOPATH is not set: %s", err)
	}

	return env, nil
}

// GoBinaryPath returns the path to the go binary to use, defined by
// GO_BINARY_PATH. This value can be a name on the PATH, or the full path
func GoBinaryPath() string {
	return envy.Get("GO_BINARY_NAME", "go")
}
