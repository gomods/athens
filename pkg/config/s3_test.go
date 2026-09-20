package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestS3EncryptionConfig(t *testing.T) {
	enabled, disabled := true, false
	tests := []struct {
		name      string
		config    string
		env       map[string]string
		algorithm string
		keyID     string
		bucketKey *bool
		wantError bool
	}{
		{name: "unset"},
		{name: "SSE-S3 from TOML", config: `ServerSideEncryption = "AES256"`, algorithm: "AES256"},
		{
			name: "SSE-KMS from TOML",
			config: `ServerSideEncryption = "aws:kms"
SSEKMSKeyID = "test-key"
BucketKeyEnabled = true`,
			algorithm: "aws:kms", keyID: "test-key", bucketKey: &enabled,
		},
		{
			name: "explicit false from TOML",
			config: `ServerSideEncryption = "aws:kms"
BucketKeyEnabled = false`,
			algorithm: "aws:kms", bucketKey: &disabled,
		},
		{
			name: "environment only",
			env: map[string]string{
				"ATHENS_S3_SERVER_SIDE_ENCRYPTION": "aws:kms",
				"ATHENS_S3_SSE_KMS_KEY_ID":         "test-key",
				"ATHENS_S3_BUCKET_KEY_ENABLED":     "true",
			},
			algorithm: "aws:kms", keyID: "test-key", bucketKey: &enabled,
		},
		{
			name: "environment overrides TOML",
			config: `ServerSideEncryption = "AES256"
SSEKMSKeyID = "original-key"
BucketKeyEnabled = true`,
			env: map[string]string{
				"ATHENS_S3_SERVER_SIDE_ENCRYPTION": "aws:kms",
				"ATHENS_S3_SSE_KMS_KEY_ID":         "replacement-key",
				"ATHENS_S3_BUCKET_KEY_ENABLED":     "false",
			},
			algorithm: "aws:kms", keyID: "replacement-key", bucketKey: &disabled,
		},
		{
			name:      "invalid bucket key boolean",
			env:       map[string]string{"ATHENS_S3_BUCKET_KEY_ENABLED": "invalid"},
			wantError: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range []string{"ATHENS_S3_SERVER_SIDE_ENCRYPTION", "ATHENS_S3_SSE_KMS_KEY_ID", "ATHENS_S3_BUCKET_KEY_ENABLED"} {
				// Register restoration before unsetting so absent and explicit false
				// are tested independently of the developer's environment.
				t.Setenv(key, "")
				require.NoError(t, os.Unsetenv(key))
			}
			for key, value := range tc.env {
				t.Setenv(key, value)
			}
			file := filepath.Join(t.TempDir(), "athens.toml")
			require.NoError(t, os.WriteFile(file, []byte(`StorageType = "s3"
[Storage.S3]
Region = "us-east-1"
Bucket = "test-bucket"
`+tc.config), 0o600))
			cfg, err := ParseConfigFile(file)
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.algorithm, cfg.Storage.S3.ServerSideEncryption)
			require.Equal(t, tc.keyID, cfg.Storage.S3.SSEKMSKeyID)
			require.Equal(t, tc.bucketKey, cfg.Storage.S3.BucketKeyEnabled)
		})
	}
}
