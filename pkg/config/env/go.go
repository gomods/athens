package env

import (
	"os"
	"strconv"

	"github.com/gobuffalo/envy"
)

// GoEnvironmentWithDefault returns environment used.
// Supported values are: 'development' and 'production'
func GoEnvironmentWithDefault(value string) string {
	return envy.Get("GO_ENV", value)
}

// GoBinPath returns the path to the go binary to use, defined by
// GO_BINARY_PATH. This value can be a name on the PATH, or the full path
func GoBinPath() string {
	return envy.Get("GO_BINARY_PATH", "go")
}

// GoGetWorkers returns how many times you can concurrently
// go mod download, this is so that low performance instances
// can manage go get more sanely and not run out of disk or memory.
func GoGetWorkers() int {
	defaultNum := 30 // 3 * cmd/go's worker count.
	str := os.Getenv("ATHENS_GOGET_WORKERS")
	if str == "" {
		return defaultNum
	}

	num, err := strconv.Atoi(str)
	if err != nil {
		return defaultNum
	}

	return num
}

// ProtocolWorkers returns how many concurrent
// requests can you handle at a time for all
// download protocol paths. This is different from
// GoGetWorkers in that you can potentially serve
// 30 requests to the Download Protocol but only 5
// at a time can stash a module from Upstream to Storage.
func ProtocolWorkers() int {
	defaultNum := 30
	str := os.Getenv("ATHENS_PROTOCOL_WORKERS")
	if str == "" {
		return defaultNum
	}

	num, err := strconv.Atoi(str)
	if err != nil {
		return defaultNum
	}

	return num
}
