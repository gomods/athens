package actions

import (
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/stretchr/testify/require"
)

func TestGetSingleFlight(t *testing.T) {
	s, err := mem.NewStorage()
	require.NoError(t, err)
	checker := storage.WithChecker(s)
	logger := log.NoOpLogger()

	t.Run("memory and empty", func(t *testing.T) {
		for _, sfType := range []string{"", "memory"} {
			w, err := getSingleFlight(logger, &config.Config{SingleFlightType: sfType}, s, checker)
			require.NoError(t, err)
			require.NotNil(t, w)
		}
	})

	t.Run("missing config errors", func(t *testing.T) {
		cases := []struct {
			name string
			conf *config.Config
		}{
			{name: "etcd nil", conf: &config.Config{SingleFlightType: "etcd"}},
			{name: "redis nil", conf: &config.Config{SingleFlightType: "redis"}},
			{name: "redis-sentinel nil", conf: &config.Config{SingleFlightType: "redis-sentinel"}},
			{name: "gcp wrong storage", conf: &config.Config{SingleFlightType: "gcp", StorageType: "memory"}},
			{name: "s3 wrong storage", conf: &config.Config{SingleFlightType: "s3", StorageType: "memory"}},
			{name: "s3 nil", conf: &config.Config{SingleFlightType: "s3", StorageType: "s3"}},
			{name: "azureblob wrong storage", conf: &config.Config{SingleFlightType: "azureblob", StorageType: "memory"}},
			{name: "unknown", conf: &config.Config{SingleFlightType: "bogus"}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := getSingleFlight(logger, tc.conf, s, checker)
				require.Error(t, err)
			})
		}
	})

	t.Run("distributed wrapper", func(t *testing.T) {
		conf := &config.Config{
			SingleFlightType: "etcd",
			SingleFlight: &config.SingleFlight{
				Etcd: &config.Etcd{Endpoints: "127.0.0.1:2379"},
			},
		}
		w, err := getSingleFlight(logger, conf, s, checker)
		require.NoError(t, err)
		require.NotNil(t, w)
	})
}
