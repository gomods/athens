package module

import (
	"context"
	"testing"

	"github.com/gobuffalo/envy"
	"github.com/stretchr/testify/assert"
)

var localCtx = context.Background()

func TestIsVersion(t *testing.T) {
	// Testing the regex
	assert.True(t, IsModVersion("v1.0.0"))
	assert.True(t, IsModVersion("v12.345.6789"))
	assert.False(t, IsModVersion("v248dadf4e9068a0b3e79f02ed0a610d935de5302"))
}

func TestPseudoversion(t *testing.T) {
	mod := "github.com/arschles/assert"
	version := "v1.0.0"
	goBinaryPath := envy.Get("GO_BINARY_PATH", "go")

	v, err := PseudoVersionFromHash(localCtx, goBinaryPath, mod, version)
	assert.NoError(t, err)
	assert.Equal(t, v, version)

	"fc2da9844984ce5093111298174706e14d4c0c47"
}
