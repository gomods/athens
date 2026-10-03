package minio

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/require"
)

func TestCatalogAdvancesPastReleaseObjects(t *testing.T) {
	objects := []minio.ObjectInfo{{Key: ".athens/toolchains/v1/a/archives/one"}, {Key: ".athens/toolchains/v1/a/archives/two"}}
	mods, cursor := fetchModsAndVersions(objects, 1)
	require.Empty(t, mods)
	require.Equal(t, objects[1].Key, cursor)
}

func TestMinIOIgnoresAmbientAWSConfiguration(t *testing.T) {
	if os.Getenv("ATHENS_MINIO_ENDPOINT") == "" {
		t.Skip("requires MinIO")
	}
	invalid := filepath.Join(t.TempDir(), "invalid-aws-config")
	require.NoError(t, os.WriteFile(invalid, []byte("[invalid configuration"), 0o600))
	t.Setenv("AWS_CONFIG_FILE", invalid)
	t.Setenv("AWS_PROFILE", "not-an-athens-profile")
	require.NotNil(t, getStorage(t))
}
