package stash

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func Test_postgresLock_lock(t *testing.T) {
	db, err := sql.Open("postgres", postgresDSN(postgresTestConfig(t)))
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	lckr := &postgresLock{db: db}
	for _, test := range lockerTests {
		t.Run(test.name, test.test(lckr))
	}
}

// TestWithPostgresLock will ensure that 5 concurrent requests will all get the first request's
// response. We can ensure that because only the first response does not return an error
// and therefore all 5 responses should have no error.
func TestWithPostgresLock(t *testing.T) {
	postgresConfig := postgresTestConfig(t)
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	ms := &mockRedisStasher{strg: strg}
	wrapper, err := WithPostgresLock(postgresConfig, storage.WithChecker(strg))
	require.NoError(t, err)
	s := wrapper(ms)

	var eg errgroup.Group
	for i := 0; i < 5; i++ {
		eg.Go(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			_, err := s.Stash(ctx, "mod", "ver")
			return err
		})
	}

	err = eg.Wait()
	if err != nil {
		t.Fatal(err)
	}
}

func postgresTestConfig(t *testing.T) *config.Postgres {
	t.Helper()
	c, err := config.Load("")
	require.NoError(t, err)
	cfg := c.Index.Postgres
	if cfg.Password == "" {
		cfg.Password = "postgres"
	}
	if os.Getenv("STATIC_PORTS") == "" {
		cfg.Host, cfg.Port = getServicePort(t, "postgres", 5432)
	}
	return cfg
}
