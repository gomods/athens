package actions

import (
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestGetSingleFlight(t *testing.T) {
	tests := []struct {
		name    string
		conf    config.Config
		wantErr string // "" means success
	}{
		{name: "default is memory", conf: config.Config{}},
		{name: "memory", conf: config.Config{SingleFlightType: "memory"}},
		{name: "etcd needs its config", conf: config.Config{SingleFlightType: "etcd"}, wantErr: "etcd config must be present"},
		{name: "redis needs its config", conf: config.Config{SingleFlightType: "redis"}, wantErr: "redis config must be present"},
		{name: "redis-sentinel needs its config", conf: config.Config{SingleFlightType: "redis-sentinel"}, wantErr: "redis config must be present"},
		{name: "s3 needs s3 storage", conf: config.Config{SingleFlightType: "s3", StorageType: "mem"}, wantErr: "only works with a s3 storage type"},
		{name: "gcp needs gcp storage", conf: config.Config{SingleFlightType: "gcp", StorageType: "mem"}, wantErr: "only works with a gcp storage type"},
		{name: "azureblob needs azureblob storage", conf: config.Config{SingleFlightType: "azureblob", StorageType: "mem"}, wantErr: "only works with a azureblob storage type"},
		{name: "unknown type", conf: config.Config{SingleFlightType: "bogus"}, wantErr: "unrecognized single flight type"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w, err := getSingleFlight(log.NoOpLogger(), &tc.conf, nil, nil)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, w)
		})
	}
}
