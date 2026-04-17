package configtest

import (
	"path/filepath"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/stretchr/testify/require"
)

// Load accepts the path to a file, constructs an absolute path to the file,
// and attempts to parse it into a Config struct.
func Load(t *testing.T, path string) *config.Config {
	t.Helper()

	absPath, err := filepath.Abs(path)
	require.NoError(t, err)

	conf, err := config.ParseFile(absPath)
	require.NoError(t, err)

	return conf
}
