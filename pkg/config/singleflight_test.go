package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestS3LockConfig(t *testing.T) {
	tests := []struct {
		name   string
		config string
		env    map[string]string
		want   *S3
	}{
		{name: "defaults", want: DefaultS3Config()},
		{
			name: "TOML",
			config: `[SingleFlight.S3]
TTL = 1200
Timeout = 30
MaxRetries = 20
`,
			want: &S3{TTL: 1200, Timeout: 30, MaxRetries: 20},
		},
		{
			name: "environment only",
			env: map[string]string{
				"ATHENS_S3_LOCK_TTL":         "1800",
				"ATHENS_S3_LOCK_TIMEOUT":     "45",
				"ATHENS_S3_LOCK_MAX_RETRIES": "25",
			},
			want: &S3{TTL: 1800, Timeout: 45, MaxRetries: 25},
		},
		{
			name: "environment overrides TOML",
			config: `[SingleFlight.S3]
TTL = 1200
Timeout = 30
MaxRetries = 20
`,
			env: map[string]string{
				"ATHENS_S3_LOCK_TTL":         "1800",
				"ATHENS_S3_LOCK_TIMEOUT":     "45",
				"ATHENS_S3_LOCK_MAX_RETRIES": "25",
			},
			want: &S3{TTL: 1800, Timeout: 45, MaxRetries: 25},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range []string{"ATHENS_SINGLE_FLIGHT_TYPE", "ATHENS_S3_LOCK_TTL", "ATHENS_S3_LOCK_TIMEOUT", "ATHENS_S3_LOCK_MAX_RETRIES"} {
				t.Setenv(key, "")
				require.NoError(t, os.Unsetenv(key))
			}
			for key, value := range tc.env {
				t.Setenv(key, value)
			}
			file := filepath.Join(t.TempDir(), "athens.toml")
			require.NoError(t, os.WriteFile(file, []byte("SingleFlightType = \"s3\"\n"+tc.config), 0o600))
			cfg, err := ParseConfigFile(file)
			require.NoError(t, err)
			require.Equal(t, "s3", cfg.SingleFlightType)
			require.Equal(t, tc.want, cfg.SingleFlight.S3)
		})
	}
}
