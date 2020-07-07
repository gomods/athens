package testconfig

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gomods/athens/internal/testutil"
	"github.com/gomods/athens/pkg/config"
	"github.com/stretchr/testify/require"
)

// LoadTestConfig loads the config file config.test.toml
func LoadTestConfig(t *testing.T) *config.Config {
	configFile := filepath.Join(testutil.AthensRoot(t), "config.test.toml")
	cfg, err := config.Load(configFile)
	require.NoError(t, err)
	return cfg
}

// ProtectedRedisConfig returns config for protectedredis.
func ProtectedRedisConfig(t *testing.T) *config.Redis {
	host, port := testutil.GetServicePort(t, "protectedredis", 6380)
	return &config.Redis{
		Endpoint: fmt.Sprintf("%s:%d", host, port),
		Password: "AthensPass1",
	}
}
