package s3

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage"
	"github.com/stretchr/testify/require"
)

func TestCatalogAdvancesPastReleaseObjects(t *testing.T) {
	objects := []types.Object{{Key: aws.String(".athens/toolchains/v1/a/archives/one")}, {Key: aws.String(".athens/toolchains/v1/a/archives/two")}}
	mods, cursor := fetchModsAndVersions(objects, 1)
	require.Empty(t, mods)
	require.Equal(t, *objects[1].Key, cursor)
	objects = append(objects, types.Object{Key: aws.String(config.PackageVersionedName("example.com/mod", "v1.0.0", "info"))})
	mods, cursor = fetchModsAndVersions(objects, 1)
	require.Len(t, mods, 1)
	require.Equal(t, *objects[2].Key, cursor)
}

func TestArchiveUsesConditionalPublicationAndEncryption(t *testing.T) {
	body := []byte("verified release")
	sum := sha256.Sum256(body)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPut, r.Method)
		require.Equal(t, "*", r.Header.Get("If-None-Match"))
		require.Equal(t, "aws:kms", r.Header.Get("X-Amz-Server-Side-Encryption"))
		require.Equal(t, "release-key", r.Header.Get("X-Amz-Server-Side-Encryption-Aws-Kms-Key-Id"))
		require.Equal(t, "true", r.Header.Get("X-Amz-Server-Side-Encryption-Bucket-Key-Enabled"))
		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.True(t, bytes.HasSuffix(data, body))
		w.Header().Set("ETag", `"saved"`)
	}))
	defer server.Close()
	enabled := true
	backend, err := New(&config.S3Config{Bucket: "releases", Key: "test", Secret: "test", Region: "us-east-1", Endpoint: server.URL, ForcePathStyle: true, ServerSideEncryption: "aws:kms", SSEKMSKeyID: "release-key", BucketKeyEnabled: &enabled}, 30*time.Second)
	require.NoError(t, err)
	info := storage.ArchiveInfo{ReleaseFile: storage.ReleaseFile{Filename: "go1.25.1.linux-amd64.tar.gz", Version: "go1.25.1", OS: "linux", Arch: "amd64", Kind: "archive", SHA256: hex.EncodeToString(sum[:]), Size: int64(len(body))}, Stable: true}
	require.NoError(t, backend.SaveArchive(t.Context(), "official", info, bytes.NewReader(body)))
}
