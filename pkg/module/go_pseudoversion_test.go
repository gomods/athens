package module

import (
	"context"
	"testing"

	"github.com/gobuffalo/envy"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

var (
	localCtx = context.Background()
	mod      = "github.com/arschles/assert"
)

func TestIsVersion(t *testing.T) {
	r := require.New(t)

	// Testing the regex
	r.True(IsSemVersion("v1.0.0"))
	r.True(IsSemVersion("v12.345.6789"))
	r.False(IsSemVersion("v248dadf4e9068a0b3e79f02ed0a610d935de5302"))
}

func TestRealVersion(t *testing.T) {
	r := require.New(t)
	goBinaryPath := envy.Get("GO_BINARY_PATH", "go")
	fs := afero.NewOsFs()
	v, err := PseudoVersionFromHash(localCtx, fs, goBinaryPath, mod, "v1.0.0")
	r.NoError(err)
	r.Equal(v, "v1.0.0")
}

func TestPseudoFromHash(t *testing.T) {
	r := require.New(t)
	version := "fc2da9844984ce5093111298174706e14d4c0c47"
	goBinaryPath := envy.Get("GO_BINARY_PATH", "go")
	fs := afero.NewOsFs()
	v, err := PseudoVersionFromHash(localCtx, fs, goBinaryPath, mod, version)
	r.NoError(err)
	r.Equal("v0.0.0-20160620175154-fc2da9844984", v)
}

func TestInvalidHash(t *testing.T) {
	r := require.New(t)
	version := "asdasdasdasdada"
	goBinaryPath := envy.Get("GO_BINARY_PATH", "go")
	fs := afero.NewOsFs()
	_, err := PseudoVersionFromHash(localCtx, fs, goBinaryPath, mod, version)
	r.Error(err)
}
