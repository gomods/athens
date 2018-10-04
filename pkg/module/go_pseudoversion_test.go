package module

import (
	"context"
	"testing"

	"github.com/gobuffalo/envy"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var (
	localCtx = context.Background()
	mod      = "github.com/arschles/assert"
)

func TestIsVersion(t *testing.T) {
	// Testing the regex
	assert.True(t, IsModVersion("v1.0.0"))
	assert.True(t, IsModVersion("v12.345.6789"))
	assert.False(t, IsModVersion("v248dadf4e9068a0b3e79f02ed0a610d935de5302"))
}

func TestRealVersion(t *testing.T) {
	goBinaryPath := envy.Get("GO_BINARY_PATH", "go")
	fs := afero.NewOsFs()
	v, err := PseudoVersionFromHash(localCtx, fs, goBinaryPath, mod, "v1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, v, version)
}

func TestPseudoFromHash(t *testing.T) {
	version := "fc2da9844984ce5093111298174706e14d4c0c47"
	goBinaryPath := envy.Get("GO_BINARY_PATH", "go")
	fs := afero.NewOsFs()
	v, err := PseudoVersionFromHash(localCtx, fs, goBinaryPath, mod, version)
	assert.NoError(t, err)
	assert.Equal(t, "v0.0.0-20160620175154-fc2da9844984", v)

}
